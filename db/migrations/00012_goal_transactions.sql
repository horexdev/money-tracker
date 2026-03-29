-- +goose Up
CREATE TABLE goal_transactions (
    id          BIGSERIAL PRIMARY KEY,
    goal_id     BIGINT NOT NULL REFERENCES savings_goals(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    type        VARCHAR(10) NOT NULL CHECK (type IN ('deposit', 'withdraw')),
    amount_cents BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_goal_transactions_goal_id ON goal_transactions(goal_id);

-- +goose Down
DROP TABLE IF EXISTS goal_transactions;
