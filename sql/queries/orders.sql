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
