-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS notify_budget_alerts       BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS notify_recurring_reminders BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS notify_weekly_summary      BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS notify_goal_milestones     BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS notify_budget_alerts,
    DROP COLUMN IF EXISTS notify_recurring_reminders,
    DROP COLUMN IF EXISTS notify_weekly_summary,
    DROP COLUMN IF EXISTS notify_goal_milestones;
