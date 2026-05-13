-- name: CreateProduct :one
INSERT INTO products (id, created_at, updated_at, name, price, category, stock, description,stock_reserved)
VALUES(
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2,
  $3,
  $4,
  $5,
  0
)
RETURNING *;

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1 FOR UPDATE;

-- name: GetAllProduct :many
SELECT * FROM products
WHERE 
  (name ILIKE '%' || @name::text || '%' OR @name::text = '') 
  AND (category = @category::text OR @category::text ='')
ORDER BY created_at ASC
LIMIT @limit_val::int
OFFSET @offset_val::int;

-- name: GetAllPrice :many
SELECT id, price, stock, name FROM products ORDER BY created_at ASC;

-- name: UpdateReservedStock :one
UPDATE products
SET stock_reserved = stock_reserved + $1
WHERE id = $2
AND stock >= (stock_reserved + $1)
RETURNING *;

-- name: UpdateFailOrder :many
UPDATE products
SET stock_reserved = stock_reserved - order_items.quantity
FROM order_items
WHERE order_items.order_id = $1 
AND order_items.product_id = products.id
AND stock >= order_items.quantity
RETURNING *;

-- name: UpdateProduct :many
UPDATE products
SET stock = stock - order_items.quantity, stock_reserved = stock_reserved - order_items.quantity
FROM order_items
WHERE order_items.order_id = $1 
AND order_items.product_id = products.id
AND stock >= order_items.quantity
RETURNING *;
