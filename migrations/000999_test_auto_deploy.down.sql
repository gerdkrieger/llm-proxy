-- Rollback for test migration
-- Removes the test comment

COMMENT ON TABLE schema_migrations IS 'Tracks applied database migrations';
