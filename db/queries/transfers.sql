-- name: CreateTransfer :one
INSERT INTO transfers (user_id, from_account_id, to_account_id, amount_cents, from_currency_code, to_currency_code, exchange_rate, note, created_at, from_tx_id, to_tx_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetTransferTxIDs :one
SELECT from_tx_id, to_tx_id FROM transfers WHERE id = $1 AND user_id = $2;

-- name: GetTransferByID :one
SELECT
    t.*,
    fa.name AS from_account_name,
    ta.name AS to_account_name
FROM transfers t
JOIN accounts fa ON fa.id = t.from_account_id
JOIN accounts ta ON ta.id = t.to_account_id
WHERE t.id = $1 AND t.user_id = $2;

-- name: ListTransfersByUser :many
SELECT
    t.*,
    fa.name AS from_account_name,
    ta.name AS to_account_name
FROM transfers t
JOIN accounts fa ON fa.id = t.from_account_id
JOIN accounts ta ON ta.id = t.to_account_id
WHERE t.user_id = $1
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTransfersByAccount :many
SELECT
    t.*,
    fa.name AS from_account_name,
    ta.name AS to_account_name
FROM transfers t
JOIN accounts fa ON fa.id = t.from_account_id
JOIN accounts ta ON ta.id = t.to_account_id
WHERE t.user_id = $1
  AND (t.from_account_id = $2 OR t.to_account_id = $2)
ORDER BY t.created_at DESC;

-- name: CountTransfersByUser :one
SELECT COUNT(*)::BIGINT FROM transfers WHERE user_id = $1;

-- name: DeleteTransfer :exec
DELETE FROM transfers WHERE id = $1 AND user_id = $2;
