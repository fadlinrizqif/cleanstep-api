-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, email, hashed_password, provider)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE email = $1;
