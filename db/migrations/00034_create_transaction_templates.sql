-- +goose Up
-- transaction_templates: user-defined presets for one-tap transaction creation.
-- Differs from recurring_transactions (no schedule); applied manually via UI.
CREATE TABLE transaction_templates (
    id            BIGSERIAL        PRIMARY KEY,
    user_id       BIGINT           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT             NOT NULL DEFAULT '',
    type          transaction_type NOT NULL,
    amount_cents  BIGINT           NOT NULL CHECK (amount_cents > 0),
    amount_fixed  BOOLEAN          NOT NULL DEFAULT TRUE,
    category_id   BIGINT           NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    account_id    BIGINT           NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    currency_code CHAR(3)          NOT NULL,
    note          TEXT             NOT NULL DEFAULT '',
    sort_order    INTEGER          NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE INDEX idx_templates_user_id    ON transaction_templates(user_id);
CREATE INDEX idx_templates_user_order ON transaction_templates(user_id, sort_order);

-- +goose Down
DROP INDEX IF EXISTS idx_templates_user_order;
DROP INDEX IF EXISTS idx_templates_user_id;
DROP TABLE IF EXISTS transaction_templates;
