CREATE TABLE IF NOT EXISTS feedback_tickets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject VARCHAR(200) NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'other',
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    model VARCHAR(200),
    error_code VARCHAR(120),
    request_id VARCHAR(200),
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_user_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_admin_at TIMESTAMPTZ,
    last_reminded_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT feedback_tickets_status_check CHECK (status IN ('open', 'pending_user', 'pending_admin', 'resolved', 'closed')),
    CONSTRAINT feedback_tickets_priority_check CHECK (priority IN ('low', 'normal', 'high', 'urgent'))
);

CREATE TABLE IF NOT EXISTS feedback_messages (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES feedback_tickets(id) ON DELETE CASCADE,
    author_type VARCHAR(16) NOT NULL,
    author_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    body TEXT NOT NULL,
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT feedback_messages_author_type_check CHECK (author_type IN ('user', 'admin'))
);

CREATE INDEX IF NOT EXISTS idx_feedback_tickets_user_updated ON feedback_tickets(user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_feedback_tickets_status_updated ON feedback_tickets(status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_feedback_tickets_unanswered ON feedback_tickets(last_user_at, last_admin_at, last_reminded_at)
    WHERE status IN ('open', 'pending_admin');
CREATE INDEX IF NOT EXISTS idx_feedback_messages_ticket_created ON feedback_messages(ticket_id, created_at ASC);
