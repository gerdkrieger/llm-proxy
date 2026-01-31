-- Migration: Add filter_matches table to track blocked content
-- This table stores detailed logs of content filtering events

CREATE TABLE IF NOT EXISTS filter_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(255) NOT NULL,  -- Request ID from request_logs
    client_id UUID REFERENCES oauth_clients(id) ON DELETE SET NULL,
    filter_id INTEGER REFERENCES content_filters(id) ON DELETE CASCADE,
    
    -- Request context
    model VARCHAR(100),
    provider VARCHAR(50),
    
    -- Filter details
    pattern TEXT NOT NULL,
    replacement TEXT NOT NULL,
    filter_type VARCHAR(20) NOT NULL,  -- word, phrase, regex
    
    -- Match details
    match_count INTEGER NOT NULL DEFAULT 1,
    matched_text TEXT,  -- Sample of what was matched (truncated for privacy)
    
    -- Metadata
    ip_address INET,
    user_agent TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_filter_matches_request_id ON filter_matches(request_id);
CREATE INDEX idx_filter_matches_client_id ON filter_matches(client_id);
CREATE INDEX idx_filter_matches_filter_id ON filter_matches(filter_id);
CREATE INDEX idx_filter_matches_created_at ON filter_matches(created_at DESC);
CREATE INDEX idx_filter_matches_provider ON filter_matches(provider);

-- Add comments
COMMENT ON TABLE filter_matches IS 'Logs of content filtering events showing blocked/replaced content';
COMMENT ON COLUMN filter_matches.matched_text IS 'Truncated sample of matched content for audit purposes';
