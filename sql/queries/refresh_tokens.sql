-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  token, 
  created_at, 
  updated_at, 
  user_id, 
  expires_at, 
  revoked_at
)
VALUES (
    $1,
    DEFAULT,
    DEFAULT,
    $2,
    $3,
    NULL
)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT user_id
FROM refresh_tokens
WHERE token = $1
AND revoked_at IS NULL
AND expires_at > now();

-- name: RevokeGivenRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now(), updated_at = now()
WHERE token = $1;
