ALTER TABLE users ADD COLUMN username VARCHAR(255) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX idx_users_username ON users(username) WHERE username <> '';