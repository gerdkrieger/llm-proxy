-- Content Filters Table
-- Stores word replacement rules for content filtering
CREATE TABLE IF NOT EXISTS content_filters (
    id SERIAL PRIMARY KEY,
    
    -- Filter details
    pattern VARCHAR(255) NOT NULL,              -- Word or phrase to match (case-insensitive)
    replacement VARCHAR(255) NOT NULL,          -- Replacement text
    description TEXT,                           -- Description of why this filter exists
    
    -- Filter type
    filter_type VARCHAR(50) NOT NULL DEFAULT 'word',  -- 'word', 'phrase', 'regex'
    case_sensitive BOOLEAN DEFAULT FALSE,       -- Whether pattern matching is case-sensitive
    
    -- Scope
    enabled BOOLEAN DEFAULT TRUE,               -- Whether filter is active
    priority INTEGER DEFAULT 0,                 -- Filter priority (higher = applied first)
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),                    -- Admin user who created this
    
    -- Statistics
    match_count INTEGER DEFAULT 0,              -- How many times this filter matched
    last_matched_at TIMESTAMP WITH TIME ZONE    -- When was this filter last triggered
);

-- Indexes
CREATE INDEX idx_content_filters_enabled ON content_filters(enabled);
CREATE INDEX idx_content_filters_priority ON content_filters(priority DESC);
CREATE INDEX idx_content_filters_pattern ON content_filters(pattern);

-- Comments
COMMENT ON TABLE content_filters IS 'Content filtering rules for word/phrase replacement in user prompts';
COMMENT ON COLUMN content_filters.pattern IS 'Pattern to match (word, phrase, or regex)';
COMMENT ON COLUMN content_filters.replacement IS 'Text to replace matches with';
COMMENT ON COLUMN content_filters.filter_type IS 'Type of filter: word, phrase, or regex';
COMMENT ON COLUMN content_filters.priority IS 'Application order (higher priority first)';
