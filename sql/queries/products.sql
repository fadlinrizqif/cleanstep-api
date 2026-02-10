-- name: CreateProduct :one
INSERT INTO products (id, created_at, updated_at, name, price, category, stock)
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

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1;

-- name: GetAllProduct :many
SELECT * FROM products ORDER BY created_at ASC;

-- name: UpdateProduct :one
UPDATE products
SET stock = stock - $1 
WHERE stock >= $1 
AND id = $2
RETURNING *;
