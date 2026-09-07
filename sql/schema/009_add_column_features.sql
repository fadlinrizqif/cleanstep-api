-- +goose Up
ALTER TABLE products ADD COLUMN features JSONB DEFAULT NULL;
-- +goose Down
ALTER TABLE products DROP COLUMN features;
