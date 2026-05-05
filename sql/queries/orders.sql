-- name: CreateOrder :one
INSERT INTO orders (id, created_at, updated_at, user_id, status, total_items)
VALUES(
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2,
  $3
)
RETURNING *;

-- name: GetOrderByID :one
SELECT id, user_id, total_items FROM orders WHERE id = $1;

-- name: UpdateStatusOrder :exec
UPDATE orders
SET status = $1
WHERE id = $2;
