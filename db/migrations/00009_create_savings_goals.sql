-- +goose Up
CREATE TABLE savings_goals (
    id            BIGSERIAL   PRIMARY KEY,
    user_id       BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT        NOT NULL,
    target_cents  BIGINT      NOT NULL CHECK (target_cents > 0),
    current_cents BIGINT      NOT NULL DEFAULT 0 CHECK (current_cents >= 0),
    currency_code CHAR(3)     NOT NULL DEFAULT 'USD',
    deadline      DATE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_savings_goals_user_id ON savings_goals(user_id);

-- +goose Down
DROP TABLE savings_goals;
