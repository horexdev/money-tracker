-- name: CreateRecurring :one
INSERT INTO recurring_transactions (user_id, account_id, category_id, type, amount_cents, currency_code, note, frequency, next_run_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetRecurringByID :one
SELECT * FROM recurring_transactions WHERE id = $1 AND user_id = $2;

-- name: ListRecurringByUser :many
SELECT
    r.*,
    c.name  AS category_name,
    c.icon AS category_icon,
    c.color AS category_color
FROM recurring_transactions r
JOIN categories c ON c.id = r.category_id
WHERE r.user_id = $1
ORDER BY r.created_at DESC;

-- name: UpdateRecurring :one
UPDATE recurring_transactions
SET account_id    = $3,
    category_id   = $4,
    type          = $5,
    amount_cents  = $6,
    currency_code = $7,
    note          = $8,
    frequency     = $9,
    next_run_at   = $10,
    updated_at    = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: ToggleRecurringActive :one
UPDATE recurring_transactions
SET is_active  = NOT is_active,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteRecurring :exec
DELETE FROM recurring_transactions WHERE id = $1 AND user_id = $2;

-- name: GetDueRecurring :many
SELECT * FROM recurring_transactions
WHERE is_active = true AND next_run_at <= $1
ORDER BY next_run_at ASC
LIMIT 100;

-- name: UpdateRecurringNextRun :exec
UPDATE recurring_transactions
SET next_run_at = $2,
    updated_at  = now()
WHERE id = $1;
