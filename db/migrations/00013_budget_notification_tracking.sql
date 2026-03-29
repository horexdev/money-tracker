-- +goose Up
ALTER TABLE budgets ADD COLUMN last_notified_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE budgets DROP COLUMN last_notified_at;
