-- +goose Up
ALTER TABLE transactions ADD COLUMN currency_code CHAR(3) NOT NULL DEFAULT 'USD';
UPDATE transactions t SET currency_code = u.currency_code FROM users u WHERE t.user_id = u.id;

-- +goose Down
ALTER TABLE transactions DROP COLUMN currency_code;
