-- name: ListUserCategories :many
SELECT * FROM categories
WHERE user_id = $1 AND deleted_at IS NULL AND type NOT IN ('transfer', 'adjustment')
ORDER BY name;

-- name: ListUserCategoriesByType :many
SELECT * FROM categories
WHERE user_id = $1
  AND deleted_at IS NULL
  AND type NOT IN ('transfer', 'adjustment')
  AND (type = $2 OR type = 'both')
ORDER BY name;

-- name: ListUserCategoriesByNameAsc :many
SELECT * FROM categories
WHERE user_id = $1 AND deleted_at IS NULL AND type NOT IN ('transfer', 'adjustment')
ORDER BY name;

-- name: ListUserCategoriesByNameDesc :many
SELECT * FROM categories
WHERE user_id = $1 AND deleted_at IS NULL AND type NOT IN ('transfer', 'adjustment')
ORDER BY name DESC;

-- name: ListUserCategoriesByTypeFilterAsc :many
SELECT * FROM categories
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (type = $2 OR type = 'both')
ORDER BY name;

-- name: ListUserCategoriesByTypeFilterDesc :many
SELECT * FROM categories
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (type = $2 OR type = 'both')
ORDER BY name DESC;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryByName :one
SELECT * FROM categories
WHERE (user_id IS NULL OR user_id = $1)
  AND LOWER(name) = LOWER($2)
  AND deleted_at IS NULL
ORDER BY user_id NULLS LAST
LIMIT 1;

-- name: GetCategoryByTypeForUser :one
SELECT * FROM categories
WHERE user_id = $1 AND type = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: GetSystemCategoryByType :one
-- Returns the system (user_id IS NULL) category of the given type.
-- Used for infrastructure categories like 'transfer' and 'adjustment'.
SELECT * FROM categories
WHERE user_id IS NULL AND type = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: HasUserCategories :one
SELECT EXISTS(
  SELECT 1 FROM categories
  WHERE user_id = $1 AND deleted_at IS NULL AND is_protected = false
) AS has_categories;

-- name: CreateUserCategory :one
INSERT INTO categories (user_id, name, emoji, type, color)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET name       = $3,
    emoji      = $4,
    type       = $5,
    color      = $6,
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
