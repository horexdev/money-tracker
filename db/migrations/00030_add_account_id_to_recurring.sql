-- +goose Up
-- Add account_id to recurring_transactions.
-- Backfill matches by currency_code; falls back to default account.
ALTER TABLE recurring_transactions
    ADD COLUMN account_id BIGINT;

-- Smart backfill: prefer account with matching currency, fall back to default.
UPDATE recurring_transactions rt
SET account_id = COALESCE(
    (SELECT a.id FROM accounts a
     WHERE a.user_id = rt.user_id AND a.currency_code = rt.currency_code
     ORDER BY a.is_default DESC, a.created_at ASC
     LIMIT 1),
    (SELECT a.id FROM accounts a
     WHERE a.user_id = rt.user_id AND a.is_default = true
     LIMIT 1)
);

-- Make NOT NULL after backfill.
ALTER TABLE recurring_transactions
    ALTER COLUMN account_id SET NOT NULL;

-- FK with RESTRICT — cannot delete account while recurring transactions reference it.
ALTER TABLE recurring_transactions
    ADD CONSTRAINT fk_recurring_account
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE RESTRICT;

CREATE INDEX idx_recurring_account_id ON recurring_transactions(account_id);

-- +goose Down
DROP INDEX IF EXISTS idx_recurring_account_id;
ALTER TABLE recurring_transactions DROP CONSTRAINT IF EXISTS fk_recurring_account;
ALTER TABLE recurring_transactions DROP COLUMN IF EXISTS account_id;
