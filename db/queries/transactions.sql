-- name: CreateTransaction :one
INSERT INTO transactions (user_id, type, amount_cents, category_id, note)
VALUES ($1, $2, $3, $4, $5)
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
    c.name  AS category_name,
    c.emoji AS category_emoji
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id = $1
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetStatsByCategory :many
SELECT
    c.name              AS category_name,
    c.emoji             AS category_emoji,
    t.type,
    SUM(t.amount_cents)::BIGINT AS total_cents,
    COUNT(*)::BIGINT    AS tx_count
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id   = $1
  AND t.created_at >= $2
  AND t.created_at <  $3
GROUP BY c.name, c.emoji, t.type
ORDER BY total_cents DESC;
