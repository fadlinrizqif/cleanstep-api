-- name: CreatePayment :one
INSERT INTO payments (
  id, 
  order_id, 
  external_id, 

  method, 
  acquirer, 
  currency, 

  status, 
  fraud_status,
  amount,

  qr_string,
  url_image,

  payload,
  
  expire_at,
  paid_at,
  created_at,
  updated_at
  )
VALUES(
  gen_random_uuid(),
  $1,
  $2,

  $3,
  $4,
  $5,

  $6,
  $7,
  $8,

  $9,
  $10,

  $11,
  
  $12,
  $13,
  NOW(),
  NOW()
)
RETURNING *;


-- name: UpdatePayment :exec
UPDATE payments 
SET status = $2, paid_at = $3
WHERE order_id = $1;

-- name: GetPaymentById :one
SELECT qr_string, url_image, expire_at FROM payments WHERE order_id = $1;
