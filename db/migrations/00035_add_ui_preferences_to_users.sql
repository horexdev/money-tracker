-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS stats_chart_style TEXT NOT NULL DEFAULT 'donut',
    ADD COLUMN IF NOT EXISTS animate_numbers   BOOLEAN;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_stats_chart_style_check,
    ADD CONSTRAINT users_stats_chart_style_check
        CHECK (stats_chart_style IN ('donut', 'stacked_bar', 'dual_bar', 'profit_bars'));

-- +goose Down
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_stats_chart_style_check;

ALTER TABLE users
    DROP COLUMN IF EXISTS stats_chart_style,
    DROP COLUMN IF EXISTS animate_numbers;
