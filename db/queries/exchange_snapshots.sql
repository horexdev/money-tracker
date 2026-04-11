-- name: UpsertExchangeRate :exec
INSERT INTO exchange_rate_snapshots (snapshot_date, base_currency, target_currency, rate)
VALUES ($1, $2, $3, $4)
ON CONFLICT (snapshot_date, base_currency, target_currency)
DO UPDATE SET rate = EXCLUDED.rate;

-- name: GetExchangeRate :one
SELECT rate
FROM exchange_rate_snapshots
WHERE snapshot_date = $1
  AND base_currency = $2
  AND target_currency = $3;

-- name: GetExchangeRateOrLatest :one
-- Returns the rate for the exact date, or the most recent rate before that date.
SELECT rate
FROM exchange_rate_snapshots
WHERE base_currency = $1
  AND target_currency = $2
  AND snapshot_date <= $3
ORDER BY snapshot_date DESC
LIMIT 1;

-- name: ListDistinctBaseCurrencies :many
-- Returns all distinct base currencies used across accounts (for daily cron).
SELECT DISTINCT currency_code
FROM accounts;

-- name: GetLatestSnapshotDate :one
SELECT COALESCE(MAX(snapshot_date), '1970-01-01'::DATE)::DATE AS latest_date
FROM exchange_rate_snapshots;
