-- name: UpsertUser :one
INSERT INTO users (id, username, first_name, last_name, currency_code)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
    SET username   = EXCLUDED.username,
        first_name = EXCLUDED.first_name,
        last_name  = EXCLUDED.last_name,
        updated_at = NOW()
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserCurrency :one
UPDATE users
SET currency_code = $2,
    updated_at    = NOW()
WHERE id = $1
RETURNING *;
