-- name: CreateTransaction :one
INSERT INTO transactions (user_id, type, amount_cents, category_id, note, currency_code, account_id, snapshot_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: CreateTransactionWithDate :one
INSERT INTO transactions (user_id, type, amount_cents, category_id, note, currency_code, created_at, account_id, snapshot_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetBalance :one
SELECT
    COALESCE(SUM(CASE WHEN type = 'income'  THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_income,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_expense
FROM transactions
WHERE user_id = $1 AND is_adjustment = false;

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
    c.name  AS category_name,
    c.icon AS category_icon,
    c.color AS category_color
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id = $1
  AND t.is_adjustment = false
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserTransactions :one
SELECT count(*)::BIGINT FROM transactions WHERE user_id = $1 AND is_adjustment = false;

-- name: GetStatsByCategory :many
SELECT
    c.id                AS category_id,
    c.name              AS category_name,
    c.icon             AS category_icon,
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
  AND t.is_adjustment = false
GROUP BY c.id, c.name, c.icon, c.color, t.type, t.currency_code
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
    c.name  AS category_name,
    c.icon AS category_icon,
    c.color AS category_color
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id      = $1
  AND t.category_id  = $2
  AND t.type         = 'expense'
  AND t.created_at  >= $3
  AND t.created_at  <  $4
  AND t.is_adjustment = false
ORDER BY t.created_at DESC;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1 AND user_id = $2;

-- name: GetBalanceByCurrency :many
SELECT
    currency_code,
    COALESCE(SUM(CASE WHEN type = 'income'  THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_income,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_expense
FROM transactions
WHERE user_id = $1 AND is_adjustment = false
GROUP BY currency_code;

-- name: GetTotalInBaseCurrency :one
-- Returns net balance (income - expense) summed across all transactions, converted to the
-- target currency using exchange_rate_snapshots. Same-currency transactions use rate 1.0.
SELECT COALESCE(SUM(
    ROUND(
        CASE WHEN t.type = 'income' THEN t.amount_cents ELSE -t.amount_cents END
        * COALESCE(ers.rate, 1.0)
    )
), 0)::BIGINT AS total_net_base_cents
FROM transactions t
LEFT JOIN exchange_rate_snapshots ers
    ON ers.snapshot_date   = t.snapshot_date
   AND ers.base_currency   = t.currency_code
   AND ers.target_currency = $2
WHERE t.user_id = $1 AND t.is_adjustment = false;

-- name: ListTransactionsByAccount :many
SELECT
    t.id,
    t.user_id,
    t.type,
    t.amount_cents,
    t.category_id,
    t.note,
    t.created_at,
    t.currency_code,
    c.name  AS category_name,
    c.icon AS category_icon,
    c.color AS category_color
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id = $1 AND t.account_id = $2
  AND t.is_adjustment = false
ORDER BY t.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountUserTransactionsByAccount :one
SELECT count(*)::BIGINT FROM transactions WHERE user_id = $1 AND account_id = $2 AND is_adjustment = false;

-- name: GetStatsByCategoryAndAccount :many
SELECT
    c.id                AS category_id,
    c.name              AS category_name,
    c.icon             AS category_icon,
    c.color             AS category_color,
    t.type,
    t.currency_code,
    SUM(t.amount_cents)::BIGINT AS total_cents,
    COUNT(*)::BIGINT    AS tx_count
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id    = $1
  AND t.account_id = $2
  AND t.created_at >= $3
  AND t.created_at <  $4
  AND t.is_adjustment = false
GROUP BY c.id, c.name, c.icon, c.color, t.type, t.currency_code
ORDER BY total_cents DESC;

-- name: GetBalanceByCurrencyAndAccount :many
SELECT
    currency_code,
    COALESCE(SUM(CASE WHEN type = 'income'  THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_income,
    COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_cents ELSE 0 END), 0)::BIGINT AS total_expense
FROM transactions
WHERE user_id = $1 AND account_id = $2 AND is_adjustment = false
GROUP BY currency_code;

-- name: UpdateTransaction :one
UPDATE transactions
SET amount_cents = $3,
    category_id  = $4,
    note         = $5,
    created_at   = $6
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: ListTransactionsWithDateRange :many
SELECT
    t.id,
    t.user_id,
    t.type,
    t.amount_cents,
    t.category_id,
    t.note,
    t.created_at,
    t.currency_code,
    t.account_id,
    c.name  AS category_name,
    c.icon AS category_icon,
    c.color AS category_color,
    a.name  AS account_name
FROM transactions t
JOIN categories c ON c.id = t.category_id
LEFT JOIN accounts a ON a.id = t.account_id
WHERE t.user_id = $1
  AND t.is_adjustment = false
  AND ($4::TIMESTAMPTZ IS NULL OR t.created_at >= $4)
  AND ($5::TIMESTAMPTZ IS NULL OR t.created_at <= $5)
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserTransactionsWithDateRange :one
SELECT count(*)::BIGINT FROM transactions
WHERE user_id = $1 AND is_adjustment = false
  AND ($2::TIMESTAMPTZ IS NULL OR created_at >= $2)
  AND ($3::TIMESTAMPTZ IS NULL OR created_at <= $3);

-- name: ListTransactionsByAccountWithDateRange :many
SELECT
    t.id,
    t.user_id,
    t.type,
    t.amount_cents,
    t.category_id,
    t.note,
    t.created_at,
    t.currency_code,
    t.account_id,
    c.name  AS category_name,
    c.icon AS category_icon,
    c.color AS category_color,
    a.name  AS account_name
FROM transactions t
JOIN categories c ON c.id = t.category_id
LEFT JOIN accounts a ON a.id = t.account_id
WHERE t.user_id = $1 AND t.account_id = $2
  AND t.is_adjustment = false
  AND ($5::TIMESTAMPTZ IS NULL OR t.created_at >= $5)
  AND ($6::TIMESTAMPTZ IS NULL OR t.created_at <= $6)
ORDER BY t.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountUserTransactionsByAccountWithDateRange :one
SELECT count(*)::BIGINT FROM transactions
WHERE user_id = $1 AND account_id = $2 AND is_adjustment = false
  AND ($3::TIMESTAMPTZ IS NULL OR created_at >= $3)
  AND ($4::TIMESTAMPTZ IS NULL OR created_at <= $4);

-- name: ListTransactionsByCategoryWithDateRange :many
SELECT
    t.id,
    t.user_id,
    t.type,
    t.amount_cents,
    t.category_id,
    t.note,
    t.created_at,
    t.currency_code,
    t.account_id,
    c.name  AS category_name,
    c.icon AS category_icon,
    c.color AS category_color,
    a.name  AS account_name
FROM transactions t
JOIN categories c ON c.id = t.category_id
LEFT JOIN accounts a ON a.id = t.account_id
WHERE t.user_id = $1 AND t.category_id = $2
  AND t.is_adjustment = false
  AND ($5::TIMESTAMPTZ IS NULL OR t.created_at >= $5)
  AND ($6::TIMESTAMPTZ IS NULL OR t.created_at <= $6)
ORDER BY t.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountUserTransactionsByCategoryWithDateRange :one
SELECT count(*)::BIGINT FROM transactions
WHERE user_id = $1 AND category_id = $2 AND is_adjustment = false
  AND ($3::TIMESTAMPTZ IS NULL OR created_at >= $3)
  AND ($4::TIMESTAMPTZ IS NULL OR created_at <= $4);

-- name: ListTransactionsByAccountAndCategoryWithDateRange :many
SELECT
    t.id,
    t.user_id,
    t.type,
    t.amount_cents,
    t.category_id,
    t.note,
    t.created_at,
    t.currency_code,
    t.account_id,
    c.name  AS category_name,
    c.icon AS category_icon,
    c.color AS category_color,
    a.name  AS account_name
FROM transactions t
JOIN categories c ON c.id = t.category_id
LEFT JOIN accounts a ON a.id = t.account_id
WHERE t.user_id = $1 AND t.account_id = $2 AND t.category_id = $3
  AND t.is_adjustment = false
  AND ($6::TIMESTAMPTZ IS NULL OR t.created_at >= $6)
  AND ($7::TIMESTAMPTZ IS NULL OR t.created_at <= $7)
ORDER BY t.created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountUserTransactionsByAccountAndCategoryWithDateRange :one
SELECT count(*)::BIGINT FROM transactions
WHERE user_id = $1 AND account_id = $2 AND category_id = $3 AND is_adjustment = false
  AND ($4::TIMESTAMPTZ IS NULL OR created_at >= $4)
  AND ($5::TIMESTAMPTZ IS NULL OR created_at <= $5);

-- name: CreateAdjustmentTransaction :one
-- Creates a balance-adjustment transaction that is hidden from history and statistics
-- but is included in balance calculations. is_adjustment is always set to true.
INSERT INTO transactions (user_id, type, amount_cents, category_id, note,
    currency_code, account_id, snapshot_date, is_adjustment)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true)
RETURNING *;
