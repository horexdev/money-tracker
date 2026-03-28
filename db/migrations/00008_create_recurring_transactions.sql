-- +goose Up
CREATE TABLE recurring_transactions (
    id            BIGSERIAL        PRIMARY KEY,
    user_id       BIGINT           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id   BIGINT           NOT NULL REFERENCES categories(id),
    type          transaction_type NOT NULL,
    amount_cents  BIGINT           NOT NULL CHECK (amount_cents > 0),
    currency_code CHAR(3)          NOT NULL DEFAULT 'USD',
    note          TEXT             NOT NULL DEFAULT '',
    frequency     VARCHAR(10)      NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
    next_run_at   TIMESTAMPTZ      NOT NULL,
    is_active     BOOLEAN          NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE INDEX idx_recurring_user_id  ON recurring_transactions(user_id);
CREATE INDEX idx_recurring_next_run ON recurring_transactions(next_run_at) WHERE is_active = true;

-- +goose Down
DROP TABLE recurring_transactions;
