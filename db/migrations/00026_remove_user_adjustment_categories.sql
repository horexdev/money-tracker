-- +goose Up

-- Remove adjustment and transfer categories that were accidentally copied to
-- users by the backfill migration (00024) before is_protected was introduced.
-- These are infrastructure-only categories and must never be user-owned.
DELETE FROM categories
WHERE user_id IS NOT NULL
  AND type IN ('adjustment', 'transfer');

-- +goose Down
-- Intentionally empty: re-inserting deleted infrastructure categories per user
-- is not meaningful; they will be re-created by the system as needed.
