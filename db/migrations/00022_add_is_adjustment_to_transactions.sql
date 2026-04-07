-- +goose Up

ALTER TABLE transactions
    ADD COLUMN is_adjustment BOOLEAN NOT NULL DEFAULT false;

-- System category used for balance adjustment transactions.
-- Type 'adjustment' hides it from user-facing category pickers.
INSERT INTO categories (user_id, name, emoji, type, color)
VALUES (NULL, 'Adjustment', '⚖️', 'adjustment', '#94a3b8')
ON CONFLICT DO NOTHING;

-- +goose Down

DELETE FROM categories WHERE user_id IS NULL AND name = 'Adjustment';

ALTER TABLE transactions DROP COLUMN is_adjustment;
