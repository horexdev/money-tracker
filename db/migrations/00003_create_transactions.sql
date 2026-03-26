-- +goose Up
CREATE TYPE transaction_type AS ENUM ('expense', 'income');

CREATE TABLE transactions (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type         transaction_type NOT NULL,
    amount_cents BIGINT           NOT NULL CHECK (amount_cents > 0),
    category_id  BIGINT           NOT NULL REFERENCES categories(id),
    note         TEXT             NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id      ON transactions(user_id);
CREATE INDEX idx_transactions_user_created ON transactions(user_id, created_at DESC);

-- +goose Down
DROP TABLE transactions;
DROP TYPE transaction_type;
