DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_cron') THEN
        PERFORM cron.unschedule('clean-tokens');
    END IF;
EXCEPTION
    WHEN undefined_function THEN
        NULL;
END $$;

DROP FUNCTION IF EXISTS clean_tokens() CASCADE;

DROP EXTENSION IF EXISTS pg_cron;

DROP TRIGGER IF EXISTS token_limit_trigger ON tokens CASCADE;

DROP FUNCTION IF EXISTS check_token_limit() CASCADE;

DROP INDEX IF EXISTS idx_tokens_user_active;
DROP INDEX IF EXISTS idx_tokens_user_created;

DROP TABLE IF EXISTS tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP INDEX IF EXISTS idx_project_files_creator_login;

DROP TABLE IF EXISTS project_files_metadata;

DROP TYPE IF EXISTS file_status;
