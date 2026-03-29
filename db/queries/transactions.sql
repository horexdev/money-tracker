-- name: CreateTransaction :one
INSERT INTO transactions (user_id, type, amount_cents, category_id, note, currency_code, exchange_rate_snapshot, base_currency_at_creation)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: CreateTransactionWithDate :one
INSERT INTO transactions (user_id, type, amount_cents, category_id, note, currency_code, exchange_rate_snapshot, base_currency_at_creation, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetBalance :one
SELECT
    COALESCE(SUM(CASE WHEN type = 'income'  THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_income,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_expense
FROM transactions
WHERE user_id = $1;

-- name: ListTransactions :many
SELECT
    t.id,
    t.user_id,
    t.type,
    t.amount_cents,
    t.category_id,
    t.note,
    t.created_at,
    t.currency_code,
    t.exchange_rate_snapshot,
    t.base_currency_at_creation,
    c.name  AS category_name,
    c.emoji AS category_emoji,
    c.color AS category_color
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id = $1
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserTransactions :one
SELECT count(*)::BIGINT FROM transactions WHERE user_id = $1;

-- name: GetStatsByCategory :many
SELECT
    c.name              AS category_name,
    c.emoji             AS category_emoji,
    c.color             AS category_color,
    t.type,
    t.currency_code,
    SUM(t.amount_cents)::BIGINT AS total_cents,
    COUNT(*)::BIGINT    AS tx_count
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id   = $1
  AND t.created_at >= $2
  AND t.created_at <  $3
GROUP BY c.name, c.emoji, t.type, t.currency_code
ORDER BY total_cents DESC;

-- name: ListTransactionsByCategoryPeriod :many
SELECT
    t.id,
    t.user_id,
    t.type,
    t.amount_cents,
    t.category_id,
    t.note,
    t.created_at,
    t.currency_code,
    t.exchange_rate_snapshot,
    t.base_currency_at_creation,
    c.name  AS category_name,
    c.emoji AS category_emoji,
    c.color AS category_color
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id      = $1
  AND t.category_id  = $2
  AND t.type         = 'expense'
  AND t.created_at  >= $3
  AND t.created_at  <  $4
ORDER BY t.created_at DESC;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1 AND user_id = $2;

-- name: GetBalanceByCurrency :many
SELECT
    currency_code,
    COALESCE(SUM(CASE WHEN type = 'income'  THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_income,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_expense
FROM transactions
WHERE user_id = $1
GROUP BY currency_code;

-- name: GetTotalInBaseCurrency :one
-- Returns net balance (income - expense) summed across all transactions converted to the user's
-- base currency using the exchange_rate_snapshot stored at the time each transaction was created.
SELECT COALESCE(
    SUM(
        CASE WHEN type = 'income' THEN amount_cents ELSE -amount_cents END
        * exchange_rate_snapshot
    ), 0
)::BIGINT AS total_net_base_cents
FROM transactions
WHERE user_id = $1;

-- name: UpdateTransaction :one
UPDATE transactions
SET amount_cents = $3,
    category_id  = $4,
    note         = $5,
    created_at   = $6
WHERE id = $1 AND user_id = $2
RETURNING *;
