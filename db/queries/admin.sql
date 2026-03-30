-- name: ListAllUsers :many
SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountAllUsers :one
SELECT count(*)::BIGINT FROM users;

-- name: CountNewUsers :one
SELECT count(*)::BIGINT FROM users WHERE created_at >= @from_ts AND created_at < @to_ts;

-- name: CountActiveUsersInPeriod :one
SELECT count(DISTINCT user_id)::BIGINT FROM transactions WHERE created_at >= @from_ts AND created_at < @to_ts;

-- name: CountRetainedUsers :one
SELECT count(*)::BIGINT
FROM users u
WHERE u.created_at >= @signup_from AND u.created_at < @signup_to
  AND EXISTS (
    SELECT 1 FROM transactions t
    WHERE t.user_id = u.id AND t.created_at >= @active_from AND t.created_at < @active_to
  );

-- name: ListAllUserIDs :many
SELECT id FROM users;
