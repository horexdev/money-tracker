-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS theme        TEXT    NOT NULL DEFAULT 'system',
    ADD COLUMN IF NOT EXISTS hide_amounts BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_theme_check,
    ADD CONSTRAINT users_theme_check CHECK (theme IN ('system', 'light', 'dark'));

-- +goose Down
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_theme_check;

ALTER TABLE users
    DROP COLUMN IF EXISTS theme,
    DROP COLUMN IF EXISTS hide_amounts;
