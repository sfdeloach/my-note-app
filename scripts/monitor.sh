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

log()    { printf '...%s\n' "$1"; }
detail() { printf '......%s\n' "$1"; }

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

echo "Health check for $(date -u '+%b %-d, %Y, %-H:%M:%S UTC')"

FAILURES=()

check_smart() {
  local health attrs raw_value attr_id label
  local -A attr_labels=(
    [5]="reallocated sectors"
    [197]="pending sectors"
    [198]="uncorrectable sectors"
  )

  log "SMART check"

  health="$(sudo "$SMARTCTL_BIN" -H "$SMART_DEVICE" 2>/dev/null | grep -i "overall-health" || true)"
  if [[ "$health" == *PASSED* ]]; then
    detail "overall health: PASSED"
  else
    detail "overall health: ${health:-UNKNOWN (no output from smartctl)}"
    FAILURES+=("SMART overall health check did not report PASSED for $SMART_DEVICE: ${health:-no output}")
  fi

  attrs="$(sudo "$SMARTCTL_BIN" -A "$SMART_DEVICE" 2>/dev/null)"
  for attr_id in 5 197 198; do
    raw_value="$(awk -v id="$attr_id" '$1 == id {print $NF}' <<< "$attrs")"
    label="${attr_labels[$attr_id]}"
    if [[ -z "$raw_value" ]]; then
      detail "$label ($attr_id): UNKNOWN (attribute not reported)"
      FAILURES+=("SMART attribute $attr_id on $SMART_DEVICE was not reported by smartctl")
    else
      detail "$label ($attr_id): $raw_value"
      if (( raw_value > 0 )); then
        FAILURES+=("SMART attribute $attr_id on $SMART_DEVICE has raw value $raw_value (expected 0)")
      fi
    fi
  done
}

check_disk() {
  local used_pct
  log "disk check"
  used_pct="$(df --output=pcent "$DISK_MOUNT" | tail -1 | tr -d ' %')"
  if [[ -z "$used_pct" ]]; then
    detail "$DISK_MOUNT usage: UNKNOWN (could not read df output)"
    FAILURES+=("Could not read disk usage for $DISK_MOUNT")
  else
    detail "$DISK_MOUNT usage: ${used_pct}% (threshold ${DISK_SPACE_THRESHOLD_PCT}%)"
    if (( used_pct >= DISK_SPACE_THRESHOLD_PCT )); then
      FAILURES+=("Disk usage on $DISK_MOUNT is ${used_pct}% (threshold ${DISK_SPACE_THRESHOLD_PCT}%)")
    fi
  fi
}

check_container() {
  local service="$1" expect="$2" container_id health
  container_id="$(docker compose ps -q "$service")"
  health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "${container_id:-x}" 2>/dev/null || echo "not-found")"
  detail "$service: $health"
  if [[ "$health" != "$expect" ]]; then
    FAILURES+=("Container '$service' status is '$health' (expected '$expect')")
  fi
}

check_smart
check_disk
log "container check"
check_container database healthy
check_container joplin-server running

if (( ${#FAILURES[@]} == 0 )); then
  curl -fsS -m 10 --retry 3 "${MONITOR_HEALTHCHECK_URL}" -o /dev/null
  echo "All health checks PASS!"
  echo "------------"
  exit 0
fi

for f in "${FAILURES[@]}"; do
  log "FAILURE: $f"
done
curl -fsS -m 10 --retry 3 --data-raw "$(printf '%s\n' "${FAILURES[@]}")" "${MONITOR_HEALTHCHECK_URL}/fail" -o /dev/null || true
echo "FAILURE! One or more of the checks did not pass."
echo "------------"
exit 1
