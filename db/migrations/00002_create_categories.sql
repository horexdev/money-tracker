-- +goose Up
CREATE TABLE categories (
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name    TEXT NOT NULL,
    emoji   TEXT NOT NULL DEFAULT '',
    UNIQUE (user_id, name)
);

-- Seed default system categories (user_id = NULL means system-wide)
INSERT INTO categories (user_id, name, emoji) VALUES
    (NULL, 'Food',          '🍔'),
    (NULL, 'Transport',     '🚌'),
    (NULL, 'Housing',       '🏠'),
    (NULL, 'Health',        '💊'),
    (NULL, 'Entertainment', '🎬'),
    (NULL, 'Shopping',      '🛍'),
    (NULL, 'Salary',        '💼'),
    (NULL, 'Other',         '📦');

-- +goose Down
DROP TABLE categories;
