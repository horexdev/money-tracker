#!/usr/bin/env bash
# Post-deploy smoke test. Verifies that the moneytracker service is active
# and stable (not crash-looping). Runs on the server via:
#   ssh user@host 'bash -s' < smoke-test.sh
set -euo pipefail

SERVICE=moneytracker
MAX_ATTEMPTS=10
SLEEP_SEC=3

echo "[smoke-test] checking ${SERVICE} service status..."

for i in $(seq 1 $MAX_ATTEMPTS); do
  if systemctl is-active --quiet "$SERVICE"; then
    echo "[smoke-test] service is active (attempt ${i}/${MAX_ATTEMPTS})"

    # Wait 2s and recheck to detect immediate crash-loops
    sleep 2
    if systemctl is-active --quiet "$SERVICE"; then
      echo "[smoke-test] service passed stability check"
      exit 0
    fi

    echo "[smoke-test] service was active but stopped — likely crash-looping"
    break
  fi

  echo "[smoke-test] not active yet (attempt ${i}/${MAX_ATTEMPTS}), waiting ${SLEEP_SEC}s..."
  sleep "$SLEEP_SEC"
done

echo "[smoke-test] ERROR: service failed to become stable"
echo "[smoke-test] last journal entries:"
journalctl -u "$SERVICE" --no-pager -n 20
exit 1
