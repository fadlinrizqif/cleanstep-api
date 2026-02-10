-- +goose Up
CREATE TABLE order_items(
  id          UUID      PRIMARY KEY,
  created_at  TIMESTAMP NOT NULL,
  updated_at  TIMESTAMP NOT NULL,
  product_id  UUID      NOT NULL,
  order_id    UUID      NOT NULL,
  quantity    INTEGER   NOT NULL,
  price       INTEGER   NOT NULL,
  CONSTRAINT  fk_product
  FOREIGN KEY (product_id)
  REFERENCES  products(id),
  CONSTRAINT  fk_order
  FOREIGN KEY (order_id)
  REFERENCES  orders(id) 
);

-- +goose Down
DROP TABLE order_items;
