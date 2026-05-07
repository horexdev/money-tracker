-- name: CreateTransactionTemplate :one
INSERT INTO transaction_templates (
    user_id, name, type, amount_cents, amount_fixed,
    category_id, account_id, currency_code, note, sort_order
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    COALESCE(
        (SELECT MAX(sort_order) + 1 FROM transaction_templates WHERE user_id = $1),
        0
    )
)
RETURNING *;

-- name: GetTransactionTemplateByID :one
SELECT
    t.*,
    c.name  AS category_name,
    c.icon  AS category_icon,
    c.color AS category_color
FROM transaction_templates t
JOIN categories c ON c.id = t.category_id
WHERE t.id = $1 AND t.user_id = $2;

-- name: ListTransactionTemplatesByUser :many
SELECT
    t.*,
    c.name  AS category_name,
    c.icon  AS category_icon,
    c.color AS category_color
FROM transaction_templates t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id = $1
ORDER BY t.sort_order ASC, t.created_at ASC;

-- name: UpdateTransactionTemplate :one
UPDATE transaction_templates
SET name          = $3,
    type          = $4,
    amount_cents  = $5,
    amount_fixed  = $6,
    category_id   = $7,
    account_id    = $8,
    currency_code = $9,
    note          = $10,
    updated_at    = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateTransactionTemplateSortOrder :execrows
UPDATE transaction_templates
SET sort_order = $3,
    updated_at = now()
WHERE id = $1 AND user_id = $2;

-- name: DeleteTransactionTemplate :execrows
DELETE FROM transaction_templates WHERE id = $1 AND user_id = $2;

-- name: CountTransactionTemplatesByAccount :one
SELECT COUNT(*) FROM transaction_templates WHERE account_id = $1;

-- name: CountTransactionTemplatesByCategory :one
SELECT COUNT(*) FROM transaction_templates WHERE category_id = $1;
