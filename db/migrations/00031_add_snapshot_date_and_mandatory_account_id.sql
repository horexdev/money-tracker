-- +goose Up

-- 1. Add snapshot_date to transactions.
ALTER TABLE transactions ADD COLUMN snapshot_date DATE;

-- Backfill from created_at (UTC calendar date).
UPDATE transactions SET snapshot_date = (created_at AT TIME ZONE 'UTC')::DATE;

ALTER TABLE transactions ALTER COLUMN snapshot_date SET NOT NULL;

-- 2. Backfill account_id where NULL: prefer account matching currency, fall back to default.
UPDATE transactions t
SET account_id = COALESCE(
    (SELECT a.id FROM accounts a
     WHERE a.user_id = t.user_id AND a.currency_code = t.currency_code
     ORDER BY a.is_default DESC, a.created_at ASC
     LIMIT 1),
    (SELECT a.id FROM accounts a
     WHERE a.user_id = t.user_id AND a.is_default = true
     LIMIT 1)
)
WHERE t.account_id IS NULL;

-- 3. Make account_id NOT NULL.
ALTER TABLE transactions ALTER COLUMN account_id SET NOT NULL;

-- 4. Replace ON DELETE SET NULL with ON DELETE RESTRICT.
-- Drop old FK (name from migration 00015).
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_account_id_fkey;
ALTER TABLE transactions
    ADD CONSTRAINT transactions_account_id_fkey
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE RESTRICT;

-- Index on snapshot_date for budget queries.
CREATE INDEX idx_transactions_snapshot_date ON transactions(snapshot_date);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_snapshot_date;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_account_id_fkey;
ALTER TABLE transactions
    ADD CONSTRAINT transactions_account_id_fkey
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE SET NULL;
ALTER TABLE transactions ALTER COLUMN account_id DROP NOT NULL;
ALTER TABLE transactions DROP COLUMN IF EXISTS snapshot_date;
