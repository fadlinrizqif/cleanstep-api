-- +goose Up
ALTER TABLE products ADD COLUMN stock_reserved INTEGER NOT NULL DEFAULT 0;
-- +goose Down
ALTER TABLE products DROP COLUMN provider;
