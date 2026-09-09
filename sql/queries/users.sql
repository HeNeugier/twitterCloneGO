-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    DEFAULT,
    DEFAULT,
    $1,
    $2
)
RETURNING *;

-- name: ClearDatabase :exec
DELETE FROM users;

-- name: ReturnUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: UpdateEmailAndHashedPassword :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = now()
WHERE id = $1
RETURNING *;
