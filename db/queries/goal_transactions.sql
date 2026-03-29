-- name: InsertGoalTransaction :exec
INSERT INTO goal_transactions (goal_id, user_id, type, amount_cents)
VALUES ($1, $2, $3, $4);

-- name: ListGoalTransactions :many
SELECT id, goal_id, user_id, type, amount_cents, created_at
FROM goal_transactions
WHERE goal_id = $1 AND user_id = $2
ORDER BY created_at DESC;
