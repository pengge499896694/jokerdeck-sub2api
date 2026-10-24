-- A caller may reuse an x-request-id across retries. Conversation logs are an
-- audit trail and must retain every request, so request_id is intentionally
-- not unique.
ALTER TABLE conversation_logs DROP CONSTRAINT IF EXISTS conversation_logs_request_id_key;
CREATE INDEX IF NOT EXISTS idx_conversation_logs_request_id ON conversation_logs (request_id);
