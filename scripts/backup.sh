#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$REPO_ROOT/.env"

log()  { printf '...%s\n' "$1"; }
fail() { log "ERROR: $1"; exit 1; }

echo "Backup for $(date -u '+%b %-d, %Y, %-H:%M:%S UTC')"

TMP_DIR=""
finish() {
  local exit_code=$?
  [[ -n "$TMP_DIR" ]] && rm -rf "$TMP_DIR"
  if (( exit_code == 0 )); then
    echo "Backup SUCCESSFUL!"
  else
    echo "Backup FAILED!"
    if [[ -n "${BACKUP_HEALTHCHECK_URL:-}" ]]; then
      curl -fsS -m 10 --retry 3 "${BACKUP_HEALTHCHECK_URL}/fail" -o /dev/null || true
    fi
  fi
  echo "------------"
  exit "$exit_code"
}
trap finish EXIT

for cmd in docker gpg aws curl; do
  command -v "$cmd" >/dev/null 2>&1 || fail "required command '$cmd' not found"
done

[[ -f "$ENV_FILE" ]] || fail "$ENV_FILE not found"

set -a
source "$ENV_FILE"
set +a

: "${BACKUP_S3_BUCKET:?BACKUP_S3_BUCKET must be set in .env}"
: "${POSTGRES_USER:?POSTGRES_USER must be set in .env}"
: "${POSTGRES_DATABASE:?POSTGRES_DATABASE must be set in .env}"
: "${BACKUP_HEALTHCHECK_URL:?BACKUP_HEALTHCHECK_URL must be set in .env}"
: "${BACKUP_GPG_PASSPHRASE_FILE:?BACKUP_GPG_PASSPHRASE_FILE must be set in .env}"

[[ -f "$BACKUP_GPG_PASSPHRASE_FILE" ]] \
  || fail "passphrase file '$BACKUP_GPG_PASSPHRASE_FILE' not found"

PASS_PERMS="$(stat -c '%a' "$BACKUP_GPG_PASSPHRASE_FILE")"
[[ "$PASS_PERMS" == "600" ]] \
  || fail "$BACKUP_GPG_PASSPHRASE_FILE must be mode 600 (found $PASS_PERMS)"

cd "$REPO_ROOT"
CONTAINER_ID="$(docker compose ps -q database)"
HEALTH="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "${CONTAINER_ID:-x}" 2>/dev/null || echo "not-found")"
[[ "$HEALTH" == "healthy" ]] \
  || fail "database container is not healthy (status: $HEALTH) — run 'docker compose ps' first"

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARCHIVE_NAME="joplin-backup-${TIMESTAMP}"
TMP_DIR="$(mktemp -d)"

log "dumping database"
docker exec "$CONTAINER_ID" pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DATABASE" > "$TMP_DIR/dump.sql"
[[ -s "$TMP_DIR/dump.sql" ]] || fail "pg_dump produced an empty file"

cp "$ENV_FILE" "$TMP_DIR/.env"
tar -cf "$TMP_DIR/${ARCHIVE_NAME}.tar" -C "$TMP_DIR" dump.sql .env
rm -f "$TMP_DIR/dump.sql" "$TMP_DIR/.env"

log "encrypting archive"
gpg --batch --yes --passphrase-file "$BACKUP_GPG_PASSPHRASE_FILE" \
    --symmetric --cipher-algo AES256 \
    --output "$TMP_DIR/${ARCHIVE_NAME}.tar.gpg" "$TMP_DIR/${ARCHIVE_NAME}.tar"
rm -f "$TMP_DIR/${ARCHIVE_NAME}.tar"

S3_KEY="${ARCHIVE_NAME}.tar.gpg"
log "uploading to s3://${BACKUP_S3_BUCKET}/${S3_KEY}"
aws s3 cp --only-show-errors "$TMP_DIR/${ARCHIVE_NAME}.tar.gpg" "s3://${BACKUP_S3_BUCKET}/${S3_KEY}"

aws s3 ls "s3://${BACKUP_S3_BUCKET}/${S3_KEY}" >/dev/null 2>&1 \
  || fail "upload verification failed — object not found in S3: s3://${BACKUP_S3_BUCKET}/${S3_KEY}"
log "upload successful"

# Rotate: keep only the 3 most recent joplin-backup-* objects, delete the rest.
# Filenames are UTC timestamps in a sortable format (YYYYMMDDTHHMMSSZ), so
# lexicographic sort == chronological order — no S3 object metadata needed.
KEEP=3
mapfile -t OBJECTS < <(
  aws s3 ls "s3://${BACKUP_S3_BUCKET}/joplin-backup-" \
    | awk '{print $4}' \
    | grep -E '^joplin-backup-[0-9]{8}T[0-9]{6}Z\.tar\.gpg$' \
    | sort
)
COUNT=${#OBJECTS[@]}
if (( COUNT > KEEP )); then
  DELETE_COUNT=$(( COUNT - KEEP ))
  for (( i = 0; i < DELETE_COUNT; i++ )); do
    log "rotating out old backup: ${OBJECTS[$i]}"
    aws s3 rm --only-show-errors "s3://${BACKUP_S3_BUCKET}/${OBJECTS[$i]}"
    log "deletion successful"
  done
else
  log "no rotation needed (${COUNT}/${KEEP} backups in S3)"
fi

curl -fsS -m 10 --retry 3 "${BACKUP_HEALTHCHECK_URL}" -o /dev/null
