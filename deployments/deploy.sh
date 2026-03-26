#!/usr/bin/env bash
# Deploy script for moneytracker. Runs on the server via:
#   ssh user@host "BOT_TOKEN='...' DATABASE_URL='...' REDIS_URL='...' bash -s" < deploy.sh
set -euxo pipefail

DEPLOY_DIR=/opt/moneytracker
SERVICE=moneytracker

cd "$DEPLOY_DIR"

echo "[deploy] stopping service..."
sudo systemctl stop "$SERVICE" || true

echo "[deploy] rotating binary (keeping bot.prev for rollback)..."
if [ -f "$DEPLOY_DIR/bot" ]; then
  cp "$DEPLOY_DIR/bot" "$DEPLOY_DIR/bot.prev"
fi
mv "$DEPLOY_DIR/bot.new" "$DEPLOY_DIR/bot"
chmod +x "$DEPLOY_DIR/bot"
chmod +x "$DEPLOY_DIR/goose"

echo "[deploy] writing .env..."
cat > "$DEPLOY_DIR/.env" <<ENVEOF
BOT_TOKEN=${BOT_TOKEN}
DATABASE_URL=${DATABASE_URL}
REDIS_URL=${REDIS_URL}
LOG_LEVEL=info
MIGRATIONS_DIR=${DEPLOY_DIR}/migrations
ENVEOF
chmod 600 "$DEPLOY_DIR/.env"

echo "[deploy] updating systemd unit..."
sudo mv /tmp/moneytracker.service /etc/systemd/system/moneytracker.service
sudo systemctl daemon-reload
sudo systemctl enable "$SERVICE"

echo "[deploy] ensuring postgres and redis are running..."
docker compose -f "$DEPLOY_DIR/docker-compose.yml" up -d --remove-orphans

echo "[deploy] waiting for postgres to be healthy..."
postgres_ready() {
  local i=0
  while [ "$i" -lt 20 ]; do
    i=$((i + 1))
    if docker compose -f "$DEPLOY_DIR/docker-compose.yml" \
         exec -T postgres pg_isready -U moneytracker -d moneytracker; then
      echo "[deploy] postgres is ready (attempt ${i}/20)"
      return 0
    fi
    if [ "$i" -ge 20 ]; then
      echo "[deploy] ERROR: postgres did not become ready in time"
      return 1
    fi
    echo "[deploy] postgres not ready yet (attempt ${i}/20), waiting 2s..."
    sleep 2
  done
}
postgres_ready && true
if [ $? -ne 0 ]; then
  echo "[deploy] starting service..."
  sudo systemctl start "$SERVICE" || true
  exit 1
fi

echo "[deploy] running migrations..."
if "$DEPLOY_DIR/goose" -dir "$DEPLOY_DIR/migrations" postgres "${DATABASE_URL}" up; then
  echo "[deploy] migrations succeeded"
else
  echo "[deploy] ERROR: migrations failed — restoring previous binary"
  if [ -f "$DEPLOY_DIR/bot.prev" ]; then
    cp "$DEPLOY_DIR/bot.prev" "$DEPLOY_DIR/bot"
  fi
  sudo systemctl start "$SERVICE"
  exit 1
fi

echo "[deploy] starting service..."
sudo systemctl start "$SERVICE"

echo "[deploy] done"
