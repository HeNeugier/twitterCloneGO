-- +goose Up
CREATE TABLE users (
  id UUID PRIMARY KEY,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  email text NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE users;
