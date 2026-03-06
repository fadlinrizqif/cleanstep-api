-- +goose Up
ALTER TABLE users
ADD COLUMN provider TEXT NOT NULL DEFAULT 'manual';

-- +goose Down
ALTER TABLE users
DROP COLUMN provider;
