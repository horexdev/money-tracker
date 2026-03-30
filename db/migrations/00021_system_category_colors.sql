-- +goose Up

-- Assign distinct colors to system categories that were all set to the default #6366f1.
UPDATE categories SET color = '#f97316' WHERE user_id IS NULL AND name = 'Food';
UPDATE categories SET color = '#8b5cf6' WHERE user_id IS NULL AND name = 'Transport';
UPDATE categories SET color = '#f97316' WHERE user_id IS NULL AND name = 'Housing';
UPDATE categories SET color = '#22c55e' WHERE user_id IS NULL AND name = 'Health';
UPDATE categories SET color = '#ec4899' WHERE user_id IS NULL AND name = 'Entertainment';
UPDATE categories SET color = '#ef4444' WHERE user_id IS NULL AND name = 'Shopping';
UPDATE categories SET color = '#10b981' WHERE user_id IS NULL AND name = 'Salary';
UPDATE categories SET color = '#64748b' WHERE user_id IS NULL AND name = 'Other';
UPDATE categories SET color = '#3b82f6' WHERE user_id IS NULL AND name = 'Savings';
UPDATE categories SET color = '#6366f1' WHERE user_id IS NULL AND name = 'Transfer';

-- +goose Down
UPDATE categories SET color = '#6366f1' WHERE user_id IS NULL;
