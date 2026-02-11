-- name: CreateOrderItems :one
INSERT INTO order_items (id, created_at, updated_at, product_id, order_id, quantity, price)
VALUES(
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2,
  $3,
  $4
)
RETURNING *;
