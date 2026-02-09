-- Auth Service Schema
-- Sessions and login attempts for authentication tracking.

BEGIN;

-- sessions stores active user sessions tied to JWT tokens.
CREATE TABLE IF NOT EXISTS sessions (
    id          VARCHAR(64)  PRIMARY KEY,
    user_id     VARCHAR(64)  NOT NULL,
    token       TEXT         NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id    ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token      ON sessions (token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);

-- login_attempts records every login for auditing and brute-force protection.
CREATE TABLE IF NOT EXISTS login_attempts (
    id          VARCHAR(64)  PRIMARY KEY,
    email       VARCHAR(255) NOT NULL,
    success     BOOLEAN      NOT NULL DEFAULT FALSE,
    ip_address  VARCHAR(45)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_email      ON login_attempts (email);
CREATE INDEX IF NOT EXISTS idx_login_attempts_created_at ON login_attempts (created_at);
CREATE INDEX IF NOT EXISTS idx_login_attempts_email_time ON login_attempts (email, created_at)
    WHERE success = FALSE;

COMMIT;
