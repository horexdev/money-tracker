-- +goose Up
ALTER TABLE transactions DROP COLUMN exchange_rate_snapshot;
ALTER TABLE transactions DROP COLUMN base_currency_at_creation;

-- +goose Down
ALTER TABLE transactions ADD COLUMN exchange_rate_snapshot NUMERIC(18, 8) NOT NULL DEFAULT 1.0;
ALTER TABLE transactions ADD COLUMN base_currency_at_creation CHAR(3) NOT NULL DEFAULT 'USD';
