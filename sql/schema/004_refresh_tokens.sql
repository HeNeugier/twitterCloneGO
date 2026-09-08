-- +goose Up
CREATE TABLE refresh_tokens (
  token text PRIMARY KEY,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz
);

-- +goose Down
DROP TABLE refresh_tokens;
