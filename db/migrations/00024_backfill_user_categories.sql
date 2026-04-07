-- +goose Up

-- Copy system (non-protected) categories to every existing user so that each
-- user owns their own personal copy. After this migration, new users receive
-- their categories via the application-level seeding in ensureUser.
INSERT INTO categories (user_id, name, emoji, type, color)
SELECT u.id, sc.name, sc.emoji, sc.type, sc.color
FROM users u
CROSS JOIN (
    SELECT name, emoji, type, color
    FROM categories
    WHERE user_id IS NULL
      AND is_protected = false
      AND deleted_at IS NULL
) sc
ON CONFLICT (user_id, name) DO NOTHING;

-- +goose Down
-- Intentionally empty: rolling back would delete user category data.
