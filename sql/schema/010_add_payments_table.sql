-- +goose Up
CREATE TABLE payments(
  id UUID PRIMARY KEY,
  order_id UUID NOT NULL, 
  external_id UUID NOT NULL,

  method TEXT NOT NULL,
  acquirer TEXT NOT NULL,
  currency TEXT NOT NULL DEFAULT 'IDR',

  status TEXT NOT NULL,
  fraud_status TEXT NOT NULL,
  amount INTEGER NOT NULL,

  qr_string TEXT NOT NULL,
  url_image TEXT NOT NULL,

  payload JSONB,

  expire_at TIMESTAMP,
  paid_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,

  CONSTRAINT fk_order FOREIGN KEY(order_id) REFERENCES orders(id)
);
-- +goose Down
DROP TABLE payments;
