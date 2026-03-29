-- name: CreateBudget :one
INSERT INTO budgets (user_id, category_id, limit_cents, period, currency_code, notify_at_percent)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetBudgetByID :one
SELECT * FROM budgets WHERE id = $1 AND user_id = $2;

-- name: ListBudgetsByUser :many
SELECT
    b.*,
    c.name  AS category_name,
    c.emoji AS category_emoji,
    c.color AS category_color
FROM budgets b
JOIN categories c ON c.id = b.category_id
WHERE b.user_id = $1
ORDER BY b.created_at DESC;

-- name: UpdateBudget :one
UPDATE budgets
SET limit_cents       = $3,
    period            = $4,
    currency_code     = $5,
    notify_at_percent = $6,
    updated_at        = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteBudget :exec
DELETE FROM budgets WHERE id = $1 AND user_id = $2;

-- name: UpdateBudgetLastNotified :exec
UPDATE budgets SET last_notified_at = now() WHERE id = $1;

-- name: ListDistinctUsersWithBudgets :many
SELECT DISTINCT user_id FROM budgets;

-- name: GetBudgetByUserCategoryPeriod :one
SELECT * FROM budgets
WHERE user_id = $1 AND category_id = $2 AND period = $3;

-- name: GetSpentInPeriod :one
SELECT COALESCE(SUM(amount_cents), 0)::BIGINT AS total_spent
FROM transactions
WHERE user_id     = $1
  AND category_id = $2
  AND type         = 'expense'
  AND currency_code = $3
  AND created_at  >= $4
  AND created_at  <  $5;
