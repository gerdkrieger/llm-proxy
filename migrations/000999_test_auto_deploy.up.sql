-- Test migration for auto-deployment system
-- This migration does nothing and is safe to run
-- It will be removed after testing

-- Add a comment to the schema_migrations table (idempotent operation)
COMMENT ON TABLE schema_migrations IS 'Tracks applied database migrations - Test comment added';
