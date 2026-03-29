-- +goose Up
ALTER TABLE transactions
    ADD COLUMN account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL;

CREATE INDEX idx_transactions_account_id ON transactions(account_id);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_account_id;
ALTER TABLE transactions DROP COLUMN account_id;
