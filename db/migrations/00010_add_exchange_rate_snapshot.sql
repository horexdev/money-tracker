-- +goose Up
ALTER TABLE transactions
    ADD COLUMN exchange_rate_snapshot  NUMERIC(18, 8) NOT NULL DEFAULT 1.0,
    ADD COLUMN base_currency_at_creation CHAR(3)      NOT NULL DEFAULT 'USD';

-- Backfill: for existing transactions assume currency_code == base currency at the time
-- (rate 1:1 is safe — they were recorded before multi-currency was introduced)
UPDATE transactions SET base_currency_at_creation = currency_code;

-- +goose Down
ALTER TABLE transactions
    DROP COLUMN exchange_rate_snapshot,
    DROP COLUMN base_currency_at_creation;
