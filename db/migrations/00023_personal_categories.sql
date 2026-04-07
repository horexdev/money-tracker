-- +goose Up

-- Mark Transfer and Adjustment as protected infrastructure categories.
-- These remain user_id = NULL and must never be modified or deleted by users.
ALTER TABLE categories ADD COLUMN is_protected BOOLEAN NOT NULL DEFAULT false;

UPDATE categories
SET is_protected = true
WHERE user_id IS NULL AND type IN ('transfer', 'adjustment');

-- Introduce a dedicated 'savings' type so SavingsGoalService can locate
-- the savings category by type rather than by a hardcoded English name.
UPDATE categories
SET type = 'savings'
WHERE user_id IS NULL AND name = 'Savings';

-- +goose Down
UPDATE categories SET type = 'both' WHERE type = 'savings';
ALTER TABLE categories DROP COLUMN is_protected;
