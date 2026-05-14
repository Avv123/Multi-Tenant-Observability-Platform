#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

ports=(3000 5433 6381 8081 8082 8083 8084 8085 8086 8123 9010 9011 1026 8026 9093)
has_conflict=0

container_for_port() {
  local port="$1"
  docker ps --format '{{.Names}}|{{.Ports}}' 2>/dev/null | while IFS='|' read -r name published; do
    if [[ "${published}" == *":${port}->"* ]]; then
      echo "${name}"
      return 0
    fi
  done
}

printf "%-6s %-18s %-8s %-8s %-s\n" "port" "process" "pid" "owner" "command"
for port in "${ports[@]}"; do
  line="$(lsof -nP -iTCP:${port} -sTCP:LISTEN -F pc 2>/dev/null | awk '
    /^p/ {pid=substr($0,2)}
    /^c/ {cmd=substr($0,2); print pid "|" cmd; exit}
  ' || true)"
  if [[ -z "${line}" ]]; then
    printf "%-6s %-18s %-8s %-8s %-s\n" "${port}" "-" "-" "free" "-"
    continue
  fi

  pid="${line%%|*}"
  cmd="${line#*|}"
  owner="external"
  pulse_container="$(container_for_port "${port}" || true)"
  if [[ -n "${pulse_container}" ]]; then
    owner="${pulse_container}"
  fi
  if [[ -z "${pulse_container}" || "${pulse_container}" != pulselens-* ]]; then
    has_conflict=1
  fi
  printf "%-6s %-18s %-8s %-8s %-s\n" "${port}" "${cmd}" "${pid}" "${owner}" "$(ps -p "${pid}" -o command= 2>/dev/null || echo "${cmd}")"
done

if [[ "${has_conflict}" -ne 0 ]]; then
  echo
  echo "port check failed: one or more required ports are occupied by non-PulseLens processes" >&2
  exit 1
fi

echo
echo "port check passed"
