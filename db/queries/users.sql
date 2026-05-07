-- name: UpsertUser :one
INSERT INTO users (id, username, first_name, last_name, language)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
    SET username      = EXCLUDED.username,
        first_name    = EXCLUDED.first_name,
        last_name     = EXCLUDED.last_name,
        language      = CASE WHEN users.language = '' OR users.language IS NULL THEN EXCLUDED.language ELSE users.language END,
        updated_at    = NOW()
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

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

-- name: UpdateNotificationPreferences :one
UPDATE users
SET notify_budget_alerts       = $2,
    notify_recurring_reminders = $3,
    notify_weekly_summary      = $4,
    notify_goal_milestones     = $5,
    updated_at                 = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserTheme :one
UPDATE users
SET theme      = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserHideAmounts :one
UPDATE users
SET hide_amounts = $2,
    updated_at   = NOW()
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
