-- +goose Up
ALTER TABLE budgets
    ADD COLUMN IF NOT EXISTS notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS last_notified_percent  INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE budgets
    DROP COLUMN IF EXISTS notifications_enabled,
    DROP COLUMN IF EXISTS last_notified_percent;
