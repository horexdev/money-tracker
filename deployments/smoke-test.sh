#!/usr/bin/env bash
# Post-deploy smoke test. Verifies that bot and api services are active
# and stable (not crash-looping). Runs on the server via:
#   ssh user@host 'bash -s' < smoke-test.sh
set -euo pipefail

MAX_ATTEMPTS=10
SLEEP_SEC=3

check_service() {
  local service=$1
  echo "[smoke-test] checking ${service} service status..."

  for i in $(seq 1 $MAX_ATTEMPTS); do
    if systemctl is-active --quiet "$service"; then
      echo "[smoke-test] ${service} is active (attempt ${i}/${MAX_ATTEMPTS})"
      sleep 2
      if systemctl is-active --quiet "$service"; then
        echo "[smoke-test] ${service} passed stability check"
        return 0
      fi
      echo "[smoke-test] ${service} was active but stopped — likely crash-looping"
      break
    fi
    echo "[smoke-test] ${service} not active yet (attempt ${i}/${MAX_ATTEMPTS}), waiting ${SLEEP_SEC}s..."
    sleep "$SLEEP_SEC"
  done

  echo "[smoke-test] ERROR: ${service} failed to become stable"
  echo "[smoke-test] last journal entries:"
  journalctl -u "$service" --no-pager -n 20
  return 1
}

check_service moneytracker
check_service moneytracker-api

echo "[smoke-test] all services are running"
