#!/usr/bin/env bash
# Restore MoneyTracker PostgreSQL database from a local backup file.
# Usage: restore.sh <backup-file-or-name>
# WARNING: This will DROP the public schema and restore from the given backup.
set -euo pipefail

DEPLOY_DIR=/opt/moneytracker
BACKUP_DIR="${DEPLOY_DIR}/backups"
CONTAINER=moneytracker-postgres-1
DBNAME=moneytracker
DBUSER=moneytracker
ENV_FILE="${DEPLOY_DIR}/.env"
BOT_SERVICE=moneytracker
API_SERVICE=moneytracker-api

usage() {
  echo "Usage: $0 <backup-file-or-name>"
  echo ""
  echo "Examples:"
  echo "  $0 /opt/moneytracker/backups/moneytracker_2026-04-06_10-00.sql.gz"
  echo "  $0 moneytracker_2026-04-06_10-00.sql.gz"
  echo ""
  echo "Available local backups:"
  ls -lh "${BACKUP_DIR}"/moneytracker_*.sql.gz 2>/dev/null || echo "  (none found in ${BACKUP_DIR})"
  exit 1
}

[ $# -eq 1 ] || usage

BACKUP_FILE="$1"
# If not an absolute path, look in BACKUP_DIR
if [ ! -f "$BACKUP_FILE" ]; then
  BACKUP_FILE="${BACKUP_DIR}/${1}"
fi
if [ ! -f "$BACKUP_FILE" ]; then
  echo "[restore] ERROR: backup file not found: $1" >&2
  echo "" >&2
  usage
fi

# ── load env ──────────────────────────────────────────────────────────────────
[ -f "$ENV_FILE" ] || { echo "[restore] ERROR: .env not found at ${ENV_FILE}" >&2; exit 1; }
# shellcheck source=/dev/null
set -a; . "$ENV_FILE"; set +a

# Extract PGPASSWORD from DATABASE_URL
_no_scheme="${DATABASE_URL#postgres://}"
_no_scheme="${_no_scheme#postgresql://}"
_userinfo="${_no_scheme%%@*}"
PGPASSWORD="${_userinfo#*:}"
export PGPASSWORD

# ── confirm ───────────────────────────────────────────────────────────────────
echo "[restore] WARNING: this will DROP and recreate the public schema in database '${DBNAME}'"
echo "[restore] Backup: ${BACKUP_FILE}"
echo ""
read -r -p "[restore] Type 'yes' to confirm: " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
  echo "[restore] aborted"
  exit 0
fi

# ── stop services ─────────────────────────────────────────────────────────────
echo "[restore] stopping services..."
sudo systemctl stop "$BOT_SERVICE" || true
sudo systemctl stop "$API_SERVICE" || true

# ── drop and recreate schema ──────────────────────────────────────────────────
echo "[restore] dropping public schema..."
docker exec -T "$CONTAINER" \
  env PGPASSWORD="$PGPASSWORD" \
  psql -U "$DBUSER" -d "$DBNAME" --no-password \
  -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# ── restore ───────────────────────────────────────────────────────────────────
echo "[restore] loading backup (this may take a while)..."
gunzip -c "$BACKUP_FILE" | docker exec -T "$CONTAINER" \
  env PGPASSWORD="$PGPASSWORD" \
  psql -U "$DBUSER" -d "$DBNAME" --no-password -q

# ── restart services ──────────────────────────────────────────────────────────
echo "[restore] restarting services..."
sudo systemctl start "$BOT_SERVICE"
sudo systemctl start "$API_SERVICE"

echo "[restore] done — database restored from ${BACKUP_FILE}"
