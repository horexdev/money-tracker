-- name: ListUserCategories :many
SELECT * FROM categories
WHERE user_id IS NULL OR user_id = $1
ORDER BY user_id NULLS FIRST, name;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryByName :one
SELECT * FROM categories
WHERE (user_id IS NULL OR user_id = $1)
  AND LOWER(name) = LOWER($2)
LIMIT 1;

-- name: CreateUserCategory :one
INSERT INTO categories (user_id, name, emoji)
VALUES ($1, $2, $3)
RETURNING *;
