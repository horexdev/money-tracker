#!/usr/bin/env bash
# Удаляет hash-ассеты, на которые нет ссылок из текущего bundle И которые старше 60 дней.
# Защита двойная: ссылки из index.html + всех current JS-чанков (lazy-imports), и mtime safety margin.
# 60 дней — большой запас: активные сессии живут максимум часы, Telegram WebView перезапускается чаще.
set -euo pipefail

WEB=/opt/moneytracker/web
ASSETS="$WEB/assets"

if [ ! -f "$WEB/index.html" ] || [ ! -d "$ASSETS" ]; then
    echo "[cleanup-assets] no index.html or assets dir — nothing to do"
    exit 0
fi

TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT

# Все references на /assets/* из текущего index.html и всех current JS-чанков.
# JS-чанки могут лениво импортировать другие чанки, поэтому grep'аем и их.
grep -ohrE 'assets/[a-zA-Z0-9._-]+\.(js|css|woff2?|svg|png|webp|jpg|ico)' \
    "$WEB/index.html" "$ASSETS"/*.js 2>/dev/null \
    | sort -u > "$TMP"

# Safety guard: при пустом списке references мы бы удалили все ассеты — это явный признак
# битого/недокаченного bundle. Лучше abort, чем пытаться чинить через rm -rf.
if [ ! -s "$TMP" ]; then
    echo "[cleanup-assets] ERROR: no asset references found in current bundle — aborting" >&2
    exit 1
fi

ref_count=$(wc -l < "$TMP")
echo "[cleanup-assets] $ref_count active references in current bundle"

deleted=0
while IFS= read -r -d '' file; do
    rel="${file#"$WEB/"}"
    if ! grep -qFx "$rel" "$TMP"; then
        rm -f "$file"
        deleted=$((deleted + 1))
    fi
done < <(find "$ASSETS" -type f -mtime +60 -print0)

echo "[cleanup-assets] removed $deleted stale asset(s)"
