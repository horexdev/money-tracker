-- name: CreateAccount :one
INSERT INTO accounts (user_id, name, icon, color, type, currency_code, is_default, include_in_total)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM accounts
WHERE id = $1 AND user_id = $2;

-- name: GetDefaultAccount :one
SELECT * FROM accounts
WHERE user_id = $1 AND is_default = true
LIMIT 1;

-- name: ListAccountsByUser :many
SELECT * FROM accounts
WHERE user_id = $1
ORDER BY is_default DESC, created_at ASC;

-- name: UpdateAccount :one
UPDATE accounts
SET name             = $3,
    icon             = $4,
    color            = $5,
    type             = $6,
    include_in_total = $7,
    updated_at       = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: ClearDefaultAccounts :exec
UPDATE accounts SET is_default = false
WHERE user_id = $1 AND is_default = true;

-- name: SetAccountDefault :one
UPDATE accounts SET is_default = true, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1 AND user_id = $2;

-- name: CountAccountTransactions :one
SELECT COUNT(*)::BIGINT FROM transactions
WHERE account_id = $1 AND user_id = $2;

-- name: CountUserAccounts :one
SELECT COUNT(*)::BIGINT FROM accounts
WHERE user_id = $1;

-- name: CountAccountTransfers :one
SELECT COUNT(*)::BIGINT FROM transfers
WHERE (from_account_id = $1 OR to_account_id = $1)
  AND user_id = $2;

-- name: GetAccountBalance :one
SELECT COALESCE(
    SUM(CASE WHEN type = 'income' THEN amount_cents ELSE -amount_cents END), 0
)::BIGINT AS balance_cents
FROM transactions
WHERE account_id = $1
  AND user_id    = $2
  AND type IN ('income', 'expense');

-- name: CountAccountRecurring :one
SELECT COUNT(*)::BIGINT FROM recurring_transactions
WHERE account_id = $1 AND user_id = $2;
