-- name: UpsertUser :one
INSERT INTO users (id, username, first_name, last_name, currency_code, language)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
    SET username      = EXCLUDED.username,
        first_name    = EXCLUDED.first_name,
        last_name     = EXCLUDED.last_name,
        language      = CASE WHEN users.language = '' OR users.language IS NULL THEN EXCLUDED.language ELSE users.language END,
        currency_code = CASE WHEN users.currency_code = '' OR users.currency_code IS NULL THEN EXCLUDED.currency_code ELSE users.currency_code END,
        updated_at    = NOW()
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserCurrency :one
UPDATE users
SET currency_code = $2,
    updated_at    = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDisplayCurrencies :one
UPDATE users
SET display_currencies = $2,
    updated_at         = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserLanguage :one
UPDATE users
SET language   = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAllUserTransactions :exec
DELETE FROM transactions WHERE user_id = $1;

-- name: DeleteAllUserBudgets :exec
DELETE FROM budgets WHERE user_id = $1;

-- name: DeleteAllUserRecurring :exec
DELETE FROM recurring_transactions WHERE user_id = $1;

-- name: DeleteAllUserGoals :exec
DELETE FROM savings_goals WHERE user_id = $1;

-- name: DeleteAllUserCategories :exec
DELETE FROM categories WHERE user_id = $1;

-- name: DeleteAllUserTransfers :exec
DELETE FROM transfers WHERE user_id = $1;

-- name: DeleteAllUserAccounts :exec
DELETE FROM accounts WHERE user_id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
