-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(),
    DEFAULT,
    DEFAULT,
    $1
)
RETURNING *;

-- name: ClearDatabase :exec
DELETE FROM users;

