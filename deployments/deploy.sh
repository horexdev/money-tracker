#!/usr/bin/env bash
# Deploy script for moneytracker. Runs on the server via:
#   ssh user@host "BOT_TOKEN='...' DATABASE_URL='...' REDIS_URL='...' ... bash -s" < deploy.sh
set -euxo pipefail

DEPLOY_DIR=/opt/moneytracker
BOT_SERVICE=moneytracker
API_SERVICE=moneytracker-api

cd "$DEPLOY_DIR"

echo "[deploy] stopping services..."
sudo systemctl stop "$BOT_SERVICE" || true
sudo systemctl stop "$API_SERVICE" || true

echo "[deploy] rotating bot binary (keeping bot.prev for rollback)..."
if [ -f "$DEPLOY_DIR/bot" ]; then
  cp "$DEPLOY_DIR/bot" "$DEPLOY_DIR/bot.prev"
fi
mv "$DEPLOY_DIR/bot.new" "$DEPLOY_DIR/bot"
chmod +x "$DEPLOY_DIR/bot"

echo "[deploy] rotating api binary (keeping api.prev for rollback)..."
if [ -f "$DEPLOY_DIR/api" ]; then
  cp "$DEPLOY_DIR/api" "$DEPLOY_DIR/api.prev"
fi
mv "$DEPLOY_DIR/api.new" "$DEPLOY_DIR/api"
chmod +x "$DEPLOY_DIR/api"

chmod +x "$DEPLOY_DIR/goose"

echo "[deploy] deploying web assets..."
rm -rf "$DEPLOY_DIR/web"
if [ -d "$DEPLOY_DIR/web-dist" ]; then
  mv "$DEPLOY_DIR/web-dist" "$DEPLOY_DIR/web"
fi

echo "[deploy] saving nginx config..."
cp "$DEPLOY_DIR/nginx.conf" "$DEPLOY_DIR/nginx.conf.active"

echo "[deploy] writing .env..."
cat > "$DEPLOY_DIR/.env" <<ENVEOF
BOT_TOKEN=${BOT_TOKEN}
DATABASE_URL=${DATABASE_URL}
REDIS_URL=${REDIS_URL}
MINI_APP_URL=${MINI_APP_URL:-}
API_PORT=${API_PORT:-8080}
ALLOWED_ORIGINS=${ALLOWED_ORIGINS:-}
LOG_LEVEL=info
MIGRATIONS_DIR=${DEPLOY_DIR}/migrations
ENVEOF
chmod 600 "$DEPLOY_DIR/.env"

echo "[deploy] updating systemd units..."
sudo mv /tmp/moneytracker.service /etc/systemd/system/moneytracker.service
sudo mv /tmp/moneytracker-api.service /etc/systemd/system/moneytracker-api.service
sudo systemctl daemon-reload
sudo systemctl enable "$BOT_SERVICE"
sudo systemctl enable "$API_SERVICE"

DOMAIN=money-tracker.hrxdev.cc
CERT_PATH=/etc/letsencrypt/live/${DOMAIN}/fullchain.pem

echo "[deploy] starting postgres and redis..."
docker compose -f "$DEPLOY_DIR/docker-compose.yml" --env-file "$DEPLOY_DIR/.env" \
  up -d --remove-orphans postgres redis

if [ ! -f "$CERT_PATH" ]; then
  echo "[deploy] no SSL cert found — stopping nginx if running, then obtaining cert..."
  docker compose -f "$DEPLOY_DIR/docker-compose.yml" stop nginx 2>/dev/null || true
  docker rm -f moneytracker-nginx-1 2>/dev/null || true

  docker run --rm \
    -p 80:80 \
    -v /etc/letsencrypt:/etc/letsencrypt \
    certbot/certbot certonly \
    --standalone \
    --non-interactive --agree-tos \
    --email "${CERTBOT_EMAIL:-admin@hrxdev.cc}" \
    -d "$DOMAIN"
  echo "[deploy] certificate issued"
fi

echo "[deploy] starting nginx with full HTTPS config..."
cp "$DEPLOY_DIR/nginx.conf.active" "$DEPLOY_DIR/nginx.conf"
docker compose -f "$DEPLOY_DIR/docker-compose.yml" --env-file "$DEPLOY_DIR/.env" up -d --remove-orphans nginx certbot

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
if ! postgres_ready; then
  sudo systemctl start "$BOT_SERVICE" || true
  sudo systemctl start "$API_SERVICE" || true
  exit 1
fi

echo "[deploy] running migrations..."
if "$DEPLOY_DIR/goose" -dir "$DEPLOY_DIR/migrations" postgres "${DATABASE_URL}" up; then
  echo "[deploy] migrations succeeded"
else
  echo "[deploy] ERROR: migrations failed — restoring previous binaries"
  if [ -f "$DEPLOY_DIR/bot.prev" ]; then
    cp "$DEPLOY_DIR/bot.prev" "$DEPLOY_DIR/bot"
  fi
  if [ -f "$DEPLOY_DIR/api.prev" ]; then
    cp "$DEPLOY_DIR/api.prev" "$DEPLOY_DIR/api"
  fi
  sudo systemctl start "$BOT_SERVICE"
  sudo systemctl start "$API_SERVICE"
  exit 1
fi

echo "[deploy] starting services..."
sudo systemctl start "$BOT_SERVICE"
sudo systemctl start "$API_SERVICE"

echo "[deploy] done"
