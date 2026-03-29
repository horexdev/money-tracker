-- +goose Up
ALTER TABLE savings_goals
    ADD COLUMN account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE savings_goals DROP COLUMN account_id;
