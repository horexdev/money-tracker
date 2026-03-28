-- name: CreateSavingsGoal :one
INSERT INTO savings_goals (user_id, name, target_cents, currency_code, deadline)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSavingsGoalByID :one
SELECT * FROM savings_goals WHERE id = $1 AND user_id = $2;

-- name: ListSavingsGoalsByUser :many
SELECT * FROM savings_goals
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateSavingsGoal :one
UPDATE savings_goals
SET name          = $3,
    target_cents  = $4,
    currency_code = $5,
    deadline      = $6,
    updated_at    = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DepositToGoal :one
UPDATE savings_goals
SET current_cents = current_cents + $3,
    updated_at    = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: WithdrawFromGoal :one
UPDATE savings_goals
SET current_cents = current_cents - $3,
    updated_at    = now()
WHERE id = $1 AND user_id = $2 AND current_cents >= $3
RETURNING *;

-- name: DeleteSavingsGoal :exec
DELETE FROM savings_goals WHERE id = $1 AND user_id = $2;
