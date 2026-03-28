-- name: ListUserCategories :many
SELECT * FROM categories
WHERE (user_id IS NULL OR user_id = $1) AND deleted_at IS NULL
ORDER BY user_id NULLS FIRST, name;

-- name: ListUserCategoriesByType :many
SELECT * FROM categories
WHERE (user_id IS NULL OR user_id = $1)
  AND deleted_at IS NULL
  AND (type = $2 OR type = 'both')
ORDER BY user_id NULLS FIRST, name;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryByName :one
SELECT * FROM categories
WHERE (user_id IS NULL OR user_id = $1)
  AND LOWER(name) = LOWER($2)
  AND deleted_at IS NULL
LIMIT 1;

-- name: CreateUserCategory :one
INSERT INTO categories (user_id, name, emoji, type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET name       = $3,
    emoji      = $4,
    type       = $5,
    updated_at = now()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: CountTransactionsByCategory :one
SELECT count(*)::BIGINT FROM transactions WHERE category_id = $1;
