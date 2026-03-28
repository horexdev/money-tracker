-- +goose Up
CREATE TABLE budgets (
    id                BIGSERIAL   PRIMARY KEY,
    user_id           BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id       BIGINT      NOT NULL REFERENCES categories(id),
    limit_cents       BIGINT      NOT NULL CHECK (limit_cents > 0),
    period            VARCHAR(10) NOT NULL CHECK (period IN ('weekly', 'monthly')),
    currency_code     CHAR(3)     NOT NULL DEFAULT 'USD',
    notify_at_percent INT         NOT NULL DEFAULT 80,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, category_id, period)
);

CREATE INDEX idx_budgets_user_id ON budgets(user_id);

-- +goose Down
DROP TABLE budgets;
