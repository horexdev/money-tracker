-- +goose Up

-- Replace the standard UNIQUE(user_id, name) constraint with one that treats
-- NULL user_id values as equal. This allows ON CONFLICT (user_id, name) to work
-- correctly for system categories (user_id IS NULL), preventing duplicates when
-- migrations or seed logic runs multiple times.
--
-- Requires PostgreSQL 15+. The project already targets postgres:16.
ALTER TABLE categories DROP CONSTRAINT categories_user_id_name_key;
ALTER TABLE categories ADD CONSTRAINT categories_user_id_name_key UNIQUE NULLS NOT DISTINCT (user_id, name);

-- +goose Down

ALTER TABLE categories DROP CONSTRAINT categories_user_id_name_key;
ALTER TABLE categories ADD CONSTRAINT categories_user_id_name_key UNIQUE (user_id, name);
