CREATE TABLE IF NOT EXISTS users (
    login VARCHAR(255) PRIMARY KEY,
    password VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tokens (
    token VARCHAR(255) PRIMARY KEY,
    user_login VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_login) REFERENCES users(login) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION check_token_limit()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COUNT(*) FROM tokens
        WHERE user_login = NEW.user_login
        AND deleted_at IS NULL) > 5 THEN

        UPDATE tokens
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE token = (
            SELECT token
            FROM tokens
            WHERE user_login = NEW.user_login
            AND deleted_at IS NULL
            ORDER BY created_at ASC
            LIMIT 1
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER token_limit_trigger
AFTER INSERT ON tokens
FOR EACH ROW
EXECUTE FUNCTION check_token_limit();

CREATE INDEX IF NOT EXISTS idx_tokens_user_created
ON tokens(user_login, created_at);

CREATE INDEX IF NOT EXISTS idx_tokens_user_active
ON tokens(user_login);

CREATE EXTENSION IF NOT EXISTS pg_cron;

CREATE OR REPLACE FUNCTION clean_tokens()
RETURNS void AS $$
BEGIN
    DELETE FROM tokens
    WHERE deleted_at IS NOT NULL
    AND deleted_at < CURRENT_TIMESTAMP - INTERVAL '7 days';
END;
$$ LANGUAGE plpgsql;

SELECT cron.schedule(
    'clean-tokens',
    '0 3 * * *',
    $$CALL clean_tokens();$$
);

CREATE TYPE file_status AS ENUM (
    'active',
    'deleted',
    'processing',
    'error'
);
