-- +goose Up
ALTER TABLE users ADD COLUMN language CHAR(2) NOT NULL DEFAULT 'en';

ALTER TABLE categories ADD COLUMN type VARCHAR(10) NOT NULL DEFAULT 'both';
ALTER TABLE categories ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE categories ADD COLUMN deleted_at TIMESTAMPTZ;

-- Set appropriate types for existing system categories.
UPDATE categories SET type = 'expense' WHERE user_id IS NULL AND name IN ('Food', 'Transport', 'Housing', 'Health', 'Entertainment', 'Shopping');
UPDATE categories SET type = 'income'  WHERE user_id IS NULL AND name = 'Salary';
UPDATE categories SET type = 'both'    WHERE user_id IS NULL AND name = 'Other';

-- +goose Down
ALTER TABLE categories DROP COLUMN deleted_at;
ALTER TABLE categories DROP COLUMN updated_at;
ALTER TABLE categories DROP COLUMN type;
ALTER TABLE users DROP COLUMN language;
