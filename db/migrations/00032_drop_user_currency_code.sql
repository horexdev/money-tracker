-- +goose Up
ALTER TABLE users DROP COLUMN currency_code;

-- +goose Down
ALTER TABLE users ADD COLUMN currency_code CHAR(3) NOT NULL DEFAULT 'USD';
