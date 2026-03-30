-- +goose Up
INSERT INTO categories (user_id, name, emoji) VALUES (NULL, 'Savings', '🏦');

-- +goose Down
DELETE FROM categories WHERE user_id IS NULL AND name = 'Savings';
