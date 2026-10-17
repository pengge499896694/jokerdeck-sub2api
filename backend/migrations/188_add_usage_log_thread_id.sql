-- Codex thread identity is distinct from session-id (shared by subagents).
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS thread_id VARCHAR(255);

-- Keep the index small: non-Codex traffic and older rows have no thread ID.
CREATE INDEX IF NOT EXISTS idx_usage_logs_user_thread_created
    ON usage_logs (user_id, thread_id, created_at DESC)
    WHERE thread_id IS NOT NULL;
