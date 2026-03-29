-- +goose Up
CREATE TABLE transfers (
    id                  BIGSERIAL     PRIMARY KEY,
    user_id             BIGINT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_account_id     BIGINT        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    to_account_id       BIGINT        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount_cents        BIGINT        NOT NULL CHECK (amount_cents > 0),
    from_currency_code  CHAR(3)       NOT NULL,
    to_currency_code    CHAR(3)       NOT NULL,
    exchange_rate       NUMERIC(18,8) NOT NULL DEFAULT 1.0,
    note                TEXT          NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT transfers_different_accounts CHECK (from_account_id <> to_account_id)
);

CREATE INDEX idx_transfers_user_id      ON transfers(user_id);
CREATE INDEX idx_transfers_from_account ON transfers(from_account_id);
CREATE INDEX idx_transfers_to_account   ON transfers(to_account_id);

-- +goose Down
DROP TABLE transfers;
