-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE email = $1;
