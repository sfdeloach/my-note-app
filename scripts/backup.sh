#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$REPO_ROOT/.env"

for cmd in docker gpg aws; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "ERROR: required command '$cmd' not found." >&2; exit 1; }
done

[[ -f "$ENV_FILE" ]] || { echo "ERROR: $ENV_FILE not found." >&2; exit 1; }

set -a
source "$ENV_FILE"
set +a

: "${BACKUP_S3_BUCKET:?BACKUP_S3_BUCKET must be set in .env}"
: "${POSTGRES_USER:?POSTGRES_USER must be set in .env}"
: "${POSTGRES_DATABASE:?POSTGRES_DATABASE must be set in .env}"

cd "$REPO_ROOT"
CONTAINER_ID="$(docker compose ps -q database)"
HEALTH="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "${CONTAINER_ID:-x}" 2>/dev/null || echo "not-found")"
if [[ "$HEALTH" != "healthy" ]]; then
  echo "ERROR: database container is not healthy (status: $HEALTH). Run 'docker compose ps' first." >&2
  exit 1
fi

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARCHIVE_NAME="joplin-backup-${TIMESTAMP}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

docker exec "$CONTAINER_ID" pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DATABASE" > "$TMP_DIR/dump.sql"
[[ -s "$TMP_DIR/dump.sql" ]] || { echo "ERROR: pg_dump produced an empty file." >&2; exit 1; }

cp "$ENV_FILE" "$TMP_DIR/.env"
tar -cf "$TMP_DIR/${ARCHIVE_NAME}.tar" -C "$TMP_DIR" dump.sql .env
rm -f "$TMP_DIR/dump.sql" "$TMP_DIR/.env"

echo "Enter the backup encryption passphrase (from your password manager) when prompted."
# Stage 4 will make this non-interactive for cron; keep interactive pinentry for now.
gpg --symmetric --cipher-algo AES256 --output "$TMP_DIR/${ARCHIVE_NAME}.tar.gpg" "$TMP_DIR/${ARCHIVE_NAME}.tar"
rm -f "$TMP_DIR/${ARCHIVE_NAME}.tar"

S3_KEY="${ARCHIVE_NAME}.tar.gpg"
aws s3 cp "$TMP_DIR/${ARCHIVE_NAME}.tar.gpg" "s3://${BACKUP_S3_BUCKET}/${S3_KEY}"

aws s3 ls "s3://${BACKUP_S3_BUCKET}/${S3_KEY}" >/dev/null 2>&1 \
  || { echo "ERROR: upload verification failed — object not found in S3." >&2; exit 1; }

echo "Backup uploaded: s3://${BACKUP_S3_BUCKET}/${S3_KEY}"
