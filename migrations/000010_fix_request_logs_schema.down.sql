-- Rollback: Remove columns added by fix

DROP INDEX IF EXISTS idx_request_logs_method;
DROP INDEX IF EXISTS idx_request_logs_path;

ALTER TABLE request_logs DROP COLUMN IF EXISTS method;
ALTER TABLE request_logs DROP COLUMN IF EXISTS path;
ALTER TABLE request_logs DROP COLUMN IF EXISTS user_agent;
ALTER TABLE request_logs DROP COLUMN IF EXISTS prompt_tokens;
ALTER TABLE request_logs DROP COLUMN IF EXISTS completion_tokens;
