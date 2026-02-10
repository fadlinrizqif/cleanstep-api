-- +goose Up
CREATE TABLE orders(
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  user_id UUID NOT NULL,
  status TEXT NOT NULL,
  total_items INTEGER NOT NULL,
  CONSTRAINT fk_users
  FOREIGN KEY (user_id)
  REFERENCES users(id) 
);

-- +goose Down
DROP TABLE orders;
