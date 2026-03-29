-- +goose Up
CREATE TYPE account_type AS ENUM ('checking', 'savings', 'cash', 'credit', 'crypto');

CREATE TABLE accounts (
    id               BIGSERIAL    PRIMARY KEY,
    user_id          BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name             TEXT         NOT NULL,
    icon             TEXT         NOT NULL DEFAULT 'wallet',
    color            TEXT         NOT NULL DEFAULT '#6366f1',
    type             account_type NOT NULL DEFAULT 'checking',
    currency_code    CHAR(3)      NOT NULL DEFAULT 'USD',
    is_default       BOOLEAN      NOT NULL DEFAULT false,
    include_in_total BOOLEAN      NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);

-- Enforce one default account per user at the DB level
CREATE UNIQUE INDEX idx_accounts_user_default ON accounts(user_id)
    WHERE is_default = true;

-- +goose Down
DROP TABLE accounts;
DROP TYPE account_type;
