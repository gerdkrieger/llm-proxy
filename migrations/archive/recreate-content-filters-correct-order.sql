-- Recreate content_filters table with correct column order
BEGIN;

-- Backup existing data
CREATE TEMP TABLE content_filters_backup AS SELECT * FROM content_filters;

-- Drop old table and dependencies
DROP TABLE IF EXISTS filter_matches CASCADE;
DROP TABLE IF EXISTS content_filters CASCADE;

-- Recreate with CORRECT column order matching the Go code
CREATE TABLE content_filters (
    id SERIAL PRIMARY KEY,
    pattern TEXT NOT NULL,
    replacement TEXT,
    description TEXT,
    filter_type VARCHAR(50) NOT NULL,
    case_sensitive BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    match_count INTEGER DEFAULT 0,
    last_matched_at TIMESTAMP
);

-- Restore data from backup
INSERT INTO content_filters (
    id, pattern, replacement, description, filter_type, 
    case_sensitive, enabled, priority, created_at, updated_at,
    created_by, match_count, last_matched_at
)
SELECT 
    id, pattern, replacement, description, filter_type,
    case_sensitive, enabled, priority, created_at, updated_at,
    created_by, match_count, last_matched_at
FROM content_filters_backup;

-- Reset sequence to continue from max ID
SELECT setval('content_filters_id_seq', COALESCE((SELECT MAX(id) FROM content_filters), 1));

-- Recreate filter_matches table
CREATE TABLE IF NOT EXISTS filter_matches (
    id SERIAL PRIMARY KEY,
    filter_id INTEGER REFERENCES content_filters(id) ON DELETE CASCADE,
    matched_text TEXT,
    replaced_text TEXT,
    matched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Verify final schema
SELECT column_name, data_type, ordinal_position
FROM information_schema.columns
WHERE table_name = 'content_filters'
ORDER BY ordinal_position;

SELECT COUNT(*) as filter_count FROM content_filters;

COMMIT;
