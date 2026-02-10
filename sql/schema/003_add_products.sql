-- +goose Up
CREATE TABLE products(
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  name TEXT NOT NULL,
  price INTEGER NOT NULL,
  category TEXT NOT NULL,
  stock INTEGER NOT NULL
);

-- +goose Down
DROP TABLE products;
