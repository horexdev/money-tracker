-- +goose Up

-- Add a system "Transfer" category used for transfer debit/credit transactions.
-- Type 'transfer' hides it from user-facing category pickers.
INSERT INTO categories (user_id, name, emoji, type)
VALUES (NULL, 'Transfer', '↔️', 'transfer')
ON CONFLICT DO NOTHING;

-- Track the auto-created debit/credit transactions on each transfer so they
-- can be cleaned up when the transfer is deleted.
ALTER TABLE transfers
    ADD COLUMN IF NOT EXISTS from_tx_id BIGINT REFERENCES transactions(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS to_tx_id   BIGINT REFERENCES transactions(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE transfers
    DROP COLUMN IF EXISTS from_tx_id,
    DROP COLUMN IF EXISTS to_tx_id;
DELETE FROM categories WHERE user_id IS NULL AND name = 'Transfer';
