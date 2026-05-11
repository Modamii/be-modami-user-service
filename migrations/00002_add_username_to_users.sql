-- +goose Up
ALTER TABLE users ADD COLUMN username VARCHAR(255) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX idx_users_username ON users(username) WHERE username <> '';

-- +goose Down
DROP INDEX IF EXISTS idx_users_username;

ALTER TABLE users DROP COLUMN IF EXISTS username;
