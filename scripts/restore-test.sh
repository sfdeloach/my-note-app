#!/usr/bin/env bash
# Quarterly restore test (roadmap.md Stage 5). Spins up a throwaway Compose
# stack (compose.restore-test.yml, project name joplin-restore-test), restores
# the latest (or a chosen) S3 backup into it, and leaves it running for manual
# verification at http://<pi-address>:22301. Run `down` afterwards to tear it
# down.
#
# Deliberately restores using the .env bundled INSIDE the backup archive, not
# the live one -- a real disaster recovery only has what's in the backup, so
# this is also what proves that half of the backup is actually sufficient.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="$REPO_ROOT/compose.restore-test.yml"
WORKDIR="$REPO_ROOT/.restore-test-workdir"
PROJECT_NAME="joplin-restore-test"
HOST_PORT=22301

usage() {
  echo "Usage: $0 up [s3-key]   # restore latest backup, or a specific S3 key" >&2
  echo "       $0 down          # tear down the throwaway stack + volume" >&2
  exit 1
}

wait_healthy() {
  local container_id="$1" timeout="$2" waited=0 status
  while true; do
    status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container_id" 2>/dev/null || echo "not-found")"
    [[ "$status" == "healthy" ]] && return 0
    (( waited >= timeout )) && { echo "ERROR: timed out waiting for healthy status (last: $status)." >&2; return 1; }
    sleep 3
    waited=$(( waited + 3 ))
  done
}

cmd_down() {
  command -v docker >/dev/null 2>&1 || { echo "ERROR: docker not found." >&2; exit 1; }
  docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down -v
  rm -rf "$WORKDIR"
  echo "Throwaway stack, volume, and workdir removed."
}

cmd_up() {
  local s3_key_override="${1:-}"

  for c in docker gpg aws; do
    command -v "$c" >/dev/null 2>&1 || { echo "ERROR: required command '$c' not found." >&2; exit 1; }
  done

  local env_file="$REPO_ROOT/.env"
  [[ -f "$env_file" ]] || { echo "ERROR: $env_file not found." >&2; exit 1; }
  [[ -f "$COMPOSE_FILE" ]] || { echo "ERROR: $COMPOSE_FILE not found." >&2; exit 1; }

  # Deliberately NOT `set -a; source "$env_file"` here: that would export
  # every var in production's .env (POSTGRES_PASSWORD, APP_BASE_URL, ...)
  # into the shell, and Compose gives shell-exported vars priority over
  # --env-file during interpolation -- silently overriding the values we
  # extract from the BACKUP's .env below with production's current ones.
  # Harmless for a same-day backup, but for an older one (e.g. before a
  # password rotation) it would mean the throwaway stack tests today's
  # credentials instead of what's actually in the backup. Only pull the two
  # values this script itself needs, from production's .env specifically.
  local BACKUP_S3_BUCKET BACKUP_GPG_PASSPHRASE_FILE
  BACKUP_S3_BUCKET="$(grep -E '^BACKUP_S3_BUCKET=' "$env_file" | cut -d= -f2-)"
  BACKUP_GPG_PASSPHRASE_FILE="$(grep -E '^BACKUP_GPG_PASSPHRASE_FILE=' "$env_file" | cut -d= -f2-)"
  : "${BACKUP_S3_BUCKET:?BACKUP_S3_BUCKET must be set in .env}"
  : "${BACKUP_GPG_PASSPHRASE_FILE:?BACKUP_GPG_PASSPHRASE_FILE must be set in .env}"

  [[ -f "$BACKUP_GPG_PASSPHRASE_FILE" ]] \
    || { echo "ERROR: passphrase file '$BACKUP_GPG_PASSPHRASE_FILE' not found." >&2; exit 1; }
  local pass_perms
  pass_perms="$(stat -c '%a' "$BACKUP_GPG_PASSPHRASE_FILE")"
  [[ "$pass_perms" == "600" ]] \
    || { echo "ERROR: $BACKUP_GPG_PASSPHRASE_FILE must be mode 600 (found $pass_perms)." >&2; exit 1; }

  local start_time
  start_time=$(date +%s)
  rm -rf "$WORKDIR"
  mkdir -p "$WORKDIR"
  # Whatever exit path this takes, never leave decrypted secrets on disk.
  trap 'rm -f "$WORKDIR"/backup.tar.gpg "$WORKDIR"/backup.tar "$WORKDIR"/dump.sql "$WORKDIR"/.env "$WORKDIR"/psql-restore.log' EXIT

  local s3_key
  if [[ -n "$s3_key_override" ]]; then
    s3_key="$s3_key_override"
  else
    s3_key="$(aws s3 ls "s3://${BACKUP_S3_BUCKET}/joplin-backup-" \
      | awk '{print $4}' \
      | grep -E '^joplin-backup-[0-9]{8}T[0-9]{6}Z\.tar\.gpg$' \
      | sort | tail -1)"
  fi
  [[ -n "$s3_key" ]] || { echo "ERROR: no backup object found in s3://${BACKUP_S3_BUCKET}/." >&2; exit 1; }
  echo "Restoring from: s3://${BACKUP_S3_BUCKET}/${s3_key}"

  aws s3 cp "s3://${BACKUP_S3_BUCKET}/${s3_key}" "$WORKDIR/backup.tar.gpg"

  gpg --batch --yes --passphrase-file "$BACKUP_GPG_PASSPHRASE_FILE" \
      --decrypt --output "$WORKDIR/backup.tar" "$WORKDIR/backup.tar.gpg"
  tar -xf "$WORKDIR/backup.tar" -C "$WORKDIR" dump.sql .env

  [[ -s "$WORKDIR/dump.sql" ]] || { echo "ERROR: extracted dump.sql is empty." >&2; exit 1; }
  [[ -s "$WORKDIR/.env" ]] || { echo "ERROR: extracted .env is empty." >&2; exit 1; }

  # Joplin Server validates each request's origin against APP_BASE_URL and
  # 404s on a mismatch. The restored .env correctly points at production's
  # port (22300), but this stack is deliberately reachable on HOST_PORT
  # instead so it can coexist with production on the same Pi -- so only this
  # one value needs adjusting. APP_PORT is left untouched: it controls the
  # container-internal listen port, which must stay 22300 to match the
  # container side of compose.restore-test.yml's "22301:22300" mapping.
  sed -i -E "s|^(APP_BASE_URL=https?://[^:/]+):[0-9]+|\1:${HOST_PORT}|" "$WORKDIR/.env"

  local restored_user restored_db
  restored_user="$(grep -E '^POSTGRES_USER=' "$WORKDIR/.env" | cut -d= -f2-)"
  restored_db="$(grep -E '^POSTGRES_DATABASE=' "$WORKDIR/.env" | cut -d= -f2-)"

  echo "Starting throwaway database (project: $PROJECT_NAME)..."
  docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" --env-file "$WORKDIR/.env" up -d database
  local db_container
  db_container="$(docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" ps -q database)"
  wait_healthy "$db_container" 60

  echo "Restoring dump into throwaway database..."
  docker exec -i "$db_container" psql -U "$restored_user" -d "$restored_db" --set ON_ERROR_STOP=1 \
    < "$WORKDIR/dump.sql" > "$WORKDIR/psql-restore.log" 2>&1

  echo "Starting throwaway joplin-server..."
  docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" --env-file "$WORKDIR/.env" up -d joplin-server
  sleep 5
  local joplin_container joplin_status
  joplin_container="$(docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" ps -q joplin-server)"
  joplin_status="$(docker inspect -f '{{.State.Status}}' "$joplin_container")"
  [[ "$joplin_status" == "running" ]] \
    || { echo "ERROR: joplin-server did not stay running (status: $joplin_status). Check: docker compose -p $PROJECT_NAME -f $COMPOSE_FILE logs joplin-server" >&2; exit 1; }

  local elapsed=$(( $(date +%s) - start_time ))
  echo
  echo "Restore-test stack is up: http://<pi-address>:${HOST_PORT}"
  echo "Time to restorable state: ${elapsed}s"
  echo
  echo "Next: log in and confirm your real notes are present and correct."
  echo "When done: $0 down"
}

[[ $# -ge 1 ]] || usage
case "$1" in
  up) cmd_up "${2:-}" ;;
  down) cmd_down ;;
  *) usage ;;
esac
