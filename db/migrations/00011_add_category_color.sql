-- +goose Up
ALTER TABLE categories ADD COLUMN color VARCHAR(7) NOT NULL DEFAULT '#6366f1';

-- +goose Down
ALTER TABLE categories DROP COLUMN color;
