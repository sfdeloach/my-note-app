#!/usr/bin/env bash
# Lightweight health monitoring — roadmap.md Stage 6. Checks SMART drive
# health, disk space, and container status; reports through the same
# dead-man's-switch pattern backup.sh uses (Stage 4), on a separate
# healthchecks.io check so backup and monitoring alerts stay distinguishable.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$REPO_ROOT/.env"

# Pi-specific constants — not secrets, so kept here rather than in .env.
# See CLAUDE.md / roadmap.md Stage 6 for hardware details.
SMARTCTL_BIN=/usr/sbin/smartctl
SMART_DEVICE=/dev/sda
DISK_MOUNT=/
DISK_SPACE_THRESHOLD_PCT=85

for cmd in docker df curl; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "ERROR: required command '$cmd' not found." >&2; exit 1; }
done
[[ -x "$SMARTCTL_BIN" ]] || { echo "ERROR: $SMARTCTL_BIN not found or not executable." >&2; exit 1; }

[[ -f "$ENV_FILE" ]] || { echo "ERROR: $ENV_FILE not found." >&2; exit 1; }

set -a
source "$ENV_FILE"
set +a

: "${MONITOR_HEALTHCHECK_URL:?MONITOR_HEALTHCHECK_URL must be set in .env}"

cd "$REPO_ROOT"

FAILURES=()

check_smart() {
  local health attrs raw_value attr_id
  health="$(sudo "$SMARTCTL_BIN" -H "$SMART_DEVICE" 2>/dev/null | grep -i "overall-health" || true)"
  if [[ "$health" != *PASSED* ]]; then
    FAILURES+=("SMART overall health check did not report PASSED for $SMART_DEVICE: ${health:-no output}")
  fi

  attrs="$(sudo "$SMARTCTL_BIN" -A "$SMART_DEVICE" 2>/dev/null)"
  for attr_id in 5 197 198; do
    raw_value="$(awk -v id="$attr_id" '$1 == id {print $NF}' <<< "$attrs")"
    if [[ -n "$raw_value" && "$raw_value" -gt 0 ]]; then
      FAILURES+=("SMART attribute $attr_id on $SMART_DEVICE has raw value $raw_value (expected 0)")
    fi
  done
}

check_disk() {
  local used_pct
  used_pct="$(df --output=pcent "$DISK_MOUNT" | tail -1 | tr -d ' %')"
  if [[ -z "$used_pct" ]]; then
    FAILURES+=("Could not read disk usage for $DISK_MOUNT")
  elif (( used_pct >= DISK_SPACE_THRESHOLD_PCT )); then
    FAILURES+=("Disk usage on $DISK_MOUNT is ${used_pct}% (threshold ${DISK_SPACE_THRESHOLD_PCT}%)")
  fi
}

check_container() {
  local service="$1" expect="$2" container_id health
  container_id="$(docker compose ps -q "$service")"
  health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "${container_id:-x}" 2>/dev/null || echo "not-found")"
  if [[ "$health" != "$expect" ]]; then
    FAILURES+=("Container '$service' status is '$health' (expected '$expect')")
  fi
}

check_smart
check_disk
check_container database healthy
check_container joplin-server running

if (( ${#FAILURES[@]} == 0 )); then
  curl -fsS -m 10 --retry 3 "${MONITOR_HEALTHCHECK_URL}" -o /dev/null
  echo "Monitoring checks passed. Healthcheck ping sent."
  exit 0
fi

printf '%s\n' "${FAILURES[@]}"
curl -fsS -m 10 --retry 3 --data-raw "$(printf '%s\n' "${FAILURES[@]}")" "${MONITOR_HEALTHCHECK_URL}/fail" -o /dev/null || true
exit 1
