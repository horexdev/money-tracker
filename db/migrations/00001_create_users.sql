-- +goose Up
CREATE TABLE users (
    id            BIGINT PRIMARY KEY,
    username      TEXT        NOT NULL DEFAULT '',
    first_name    TEXT        NOT NULL DEFAULT '',
    last_name     TEXT        NOT NULL DEFAULT '',
    currency_code CHAR(3)     NOT NULL DEFAULT 'USD',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE users;
