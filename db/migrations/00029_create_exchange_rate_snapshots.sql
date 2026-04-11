-- +goose Up
CREATE TABLE exchange_rate_snapshots (
    id              BIGSERIAL    PRIMARY KEY,
    snapshot_date   DATE         NOT NULL,
    base_currency   CHAR(3)      NOT NULL,
    target_currency CHAR(3)      NOT NULL,
    rate            NUMERIC(18,8) NOT NULL CHECK (rate > 0),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE(snapshot_date, base_currency, target_currency)
);

CREATE INDEX idx_ers_date_base ON exchange_rate_snapshots(snapshot_date, base_currency);

-- +goose Down
DROP TABLE IF EXISTS exchange_rate_snapshots;
