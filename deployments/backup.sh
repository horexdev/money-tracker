#!/usr/bin/env bash
# Backup MoneyTracker PostgreSQL database to Google Drive via rclone.
# Triggered every 2 hours by moneytracker-backup.timer.
# View logs: journalctl -u moneytracker-backup.service
set -euo pipefail

DEPLOY_DIR=/opt/moneytracker
ENV_FILE="${DEPLOY_DIR}/.env"
BACKUP_DIR="${DEPLOY_DIR}/backups"
CONTAINER=moneytracker-postgres-1
DBNAME=moneytracker
DBUSER=moneytracker
RCLONE_REMOTE="${RCLONE_REMOTE:-gdrive}"
RCLONE_DEST="${RCLONE_REMOTE}:"
RETENTION_DAYS=7

# ── guards ────────────────────────────────────────────────────────────────────
command -v docker &>/dev/null || { echo "[backup] ERROR: docker not found" >&2; exit 1; }
command -v rclone &>/dev/null || { echo "[backup] ERROR: rclone not installed. Install: curl https://rclone.org/install.sh | sudo bash" >&2; exit 1; }
command -v gzip   &>/dev/null || { echo "[backup] ERROR: gzip not found" >&2; exit 1; }
[ -f "$ENV_FILE" ]            || { echo "[backup] ERROR: .env not found at ${ENV_FILE}" >&2; exit 1; }
[ -f "${DEPLOY_DIR}/rclone.conf" ] || {
  echo "[backup] ERROR: rclone.conf not found at ${DEPLOY_DIR}/rclone.conf" >&2
  echo "[backup] Copy rclone.conf.example to rclone.conf and fill in your credentials." >&2
  exit 1
}

# ── load env ──────────────────────────────────────────────────────────────────
# shellcheck source=/dev/null
set -a; . "$ENV_FILE"; set +a

# Extract PGPASSWORD from DATABASE_URL=postgres://user:pass@host:port/dbname
_no_scheme="${DATABASE_URL#postgres://}"
_no_scheme="${_no_scheme#postgresql://}"
_userinfo="${_no_scheme%%@*}"    # user:pass
PGPASSWORD="${_userinfo#*:}"     # pass only
export PGPASSWORD

# ── prepare ───────────────────────────────────────────────────────────────────
mkdir -p "$BACKUP_DIR"
TIMESTAMP=$(date '+%Y-%m-%d_%H-%M')
FILENAME="moneytracker_${TIMESTAMP}.sql.gz"
FILEPATH="${BACKUP_DIR}/${FILENAME}"

echo "[backup] starting backup: ${FILENAME}"

# ── pg_dump ───────────────────────────────────────────────────────────────────
if ! docker exec "$CONTAINER" \
    env PGPASSWORD="$PGPASSWORD" \
    pg_dump -U "$DBUSER" -d "$DBNAME" --no-password \
    | gzip -9 > "$FILEPATH"; then
  echo "[backup] ERROR: pg_dump or gzip failed" >&2
  rm -f "$FILEPATH"
  exit 1
fi

# ── integrity check ───────────────────────────────────────────────────────────
FILESIZE=$(stat -c%s "$FILEPATH" 2>/dev/null || echo 0)
if [ "$FILESIZE" -lt 100 ]; then
  echo "[backup] ERROR: backup file suspiciously small (${FILESIZE} bytes), aborting upload" >&2
  rm -f "$FILEPATH"
  exit 1
fi
echo "[backup] created: ${FILEPATH} (${FILESIZE} bytes)"

# ── upload to Google Drive ────────────────────────────────────────────────────
echo "[backup] uploading to ${RCLONE_DEST}/"
if rclone copy "$FILEPATH" "${RCLONE_DEST}/" \
    --config "${DEPLOY_DIR}/rclone.conf" \
    --log-level INFO; then
  echo "[backup] upload succeeded"
else
  echo "[backup] WARNING: upload to Google Drive failed — local backup retained at ${FILEPATH}" >&2
fi

# ── rotate old local backups ──────────────────────────────────────────────────
echo "[backup] rotating local backups older than ${RETENTION_DAYS} days..."
find "$BACKUP_DIR" -name 'moneytracker_*.sql.gz' -mtime +"$RETENTION_DAYS" -delete
echo "[backup] done"
