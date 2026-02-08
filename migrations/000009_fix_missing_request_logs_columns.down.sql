-- Rollback: Remove request_id and ip_address columns

DROP INDEX IF EXISTS idx_request_logs_request_id;
DROP INDEX IF EXISTS idx_request_logs_ip;

ALTER TABLE request_logs DROP COLUMN IF EXISTS request_id;
ALTER TABLE request_logs DROP COLUMN IF EXISTS ip_address;
