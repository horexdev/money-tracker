-- name: CreateBudget :one
INSERT INTO budgets (user_id, category_id, limit_cents, period, currency_code, notify_at_percent, notifications_enabled)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetBudgetByID :one
SELECT * FROM budgets WHERE id = $1 AND user_id = $2;

-- name: ListBudgetsByUser :many
SELECT
    b.*,
    c.name  AS category_name,
    c.icon AS category_icon,
    c.color AS category_color
FROM budgets b
JOIN categories c ON c.id = b.category_id
WHERE b.user_id = $1
ORDER BY b.created_at DESC;

-- name: UpdateBudget :one
UPDATE budgets
SET limit_cents             = $3,
    period                  = $4,
    currency_code           = $5,
    notify_at_percent       = $6,
    notifications_enabled   = $7,
    updated_at              = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteBudget :exec
DELETE FROM budgets WHERE id = $1 AND user_id = $2;

-- name: UpdateBudgetLastNotified :exec
UPDATE budgets
SET last_notified_at     = now(),
    last_notified_percent = $2
WHERE id = $1;

-- name: ListDistinctUsersWithBudgets :many
SELECT DISTINCT user_id FROM budgets;

-- name: GetBudgetByUserCategoryPeriod :one
SELECT * FROM budgets
WHERE user_id = $1 AND category_id = $2 AND period = $3;

-- name: GetSpentInPeriod :one
-- Cross-currency aggregation: converts each transaction's amount to the
-- budget's target currency using exchange_rate_snapshots. Same-currency
-- transactions pass through at rate 1.0 (no snapshot row needed).
SELECT COALESCE(SUM(
    ROUND(t.amount_cents * COALESCE(ers.rate, 1.0))
), 0)::BIGINT AS total_spent
FROM transactions t
LEFT JOIN exchange_rate_snapshots ers
    ON ers.snapshot_date   = t.snapshot_date
   AND ers.base_currency   = t.currency_code
   AND ers.target_currency = $3
WHERE t.user_id     = $1
  AND t.category_id = $2
  AND t.type         = 'expense'
  AND t.created_at  >= $4
  AND t.created_at  <  $5;
