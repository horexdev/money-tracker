-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS theme        TEXT    NOT NULL DEFAULT 'system'
        CHECK (theme IN ('system','light','dark')),
    ADD COLUMN IF NOT EXISTS hide_amounts BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS theme,
    DROP COLUMN IF EXISTS hide_amounts;
