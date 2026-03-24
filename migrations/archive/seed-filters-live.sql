-- Seed default content filters for LLM-Proxy
-- This script populates the content_filters table with predefined filters

-- Clear existing filters (optional - comment out if you want to keep existing)
-- DELETE FROM content_filters;

-- Insert default filters (without case_sensitive column for LIVE compatibility)
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('badword', 'word', '[FILTERED]', 100, 'Filtert offensive Sprache', true, 'profanity');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('damn', 'word', '[*]', 100, 'Milde Kraftausdrücke', true, 'profanity');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('shit', 'word', '[CENSORED]', 100, 'Starke Kraftausdrücke', true, 'profanity');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('confidential information', 'phrase', '[REDACTED]', 95, 'Vertrauliche Informationen', true, 'security');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('Project Phoenix', 'phrase', '[INTERNAL_PROJECT]', 95, 'Interner Projektname', true, 'internal');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('top secret', 'phrase', '[CLASSIFIED]', 95, 'Geheime Informationen', true, 'security');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES (E'\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}\\b', 'regex', '[EMAIL]', 90, 'Email-Adressen filtern', true, 'pii');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES (E'\\b0[0-9]{2,4}[-\\s]?[0-9]{3,8}\\b', 'regex', '[PHONE]', 90, 'Deutsche Telefonnummern', true, 'pii');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES (E'\\b[0-9]{4}[\\s-]?[0-9]{4}[\\s-]?[0-9]{4}[\\s-]?[0-9]{4}\\b', 'regex', '[CREDIT_CARD]', 95, 'Kreditkartennummern', true, 'pii');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('CompetitorX', 'word', '[COMPETITOR]', 80, 'Konkurrenz-Erwähnung', true, 'business');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('password', 'word', '[***]', 100, 'Passwort-Erwähnung', true, 'security');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('secret key', 'phrase', '[KEY_REDACTED]', 95, 'API Keys', true, 'security');
INSERT INTO content_filters (pattern, filter_type, replacement, priority, description, enabled, category) VALUES ('email@example.com', 'word', '[EMAIL]', 100, 'Email ', true, 'pii');

SELECT 'Imported ' || COUNT(*) || ' filters' as result FROM content_filters;
