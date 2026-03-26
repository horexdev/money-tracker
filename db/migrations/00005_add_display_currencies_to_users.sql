-- +goose Up
ALTER TABLE users ADD COLUMN display_currencies TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN display_currencies;
