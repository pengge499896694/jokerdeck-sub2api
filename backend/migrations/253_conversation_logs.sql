CREATE TABLE IF NOT EXISTS conversation_logs (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(128) NOT NULL,
    user_id BIGINT NOT NULL,
    user_email VARCHAR(255) NOT NULL DEFAULT '',
    api_key_id BIGINT NOT NULL,
    api_key_name VARCHAR(255) NOT NULL DEFAULT '',
    group_id BIGINT,
    group_name VARCHAR(255) NOT NULL DEFAULT '',
    model VARCHAR(255) NOT NULL DEFAULT '',
    endpoint VARCHAR(512) NOT NULL DEFAULT '',
    stream BOOLEAN NOT NULL DEFAULT FALSE,
    status_code INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    question TEXT NOT NULL DEFAULT '',
    answer TEXT NOT NULL DEFAULT '',
    question_bytes INTEGER NOT NULL DEFAULT 0,
    answer_bytes INTEGER NOT NULL DEFAULT 0,
    question_truncated BOOLEAN NOT NULL DEFAULT FALSE,
    answer_truncated BOOLEAN NOT NULL DEFAULT FALSE,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT conversation_logs_question_size CHECK (octet_length(question) <= 32768),
    CONSTRAINT conversation_logs_answer_size CHECK (octet_length(answer) <= 65536)
);

CREATE INDEX IF NOT EXISTS idx_conversation_logs_created_id ON conversation_logs (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_logs_request_id ON conversation_logs (request_id);
CREATE INDEX IF NOT EXISTS idx_conversation_logs_user_created ON conversation_logs (user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_logs_key_created ON conversation_logs (api_key_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_logs_group_created ON conversation_logs (group_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_logs_model_created ON conversation_logs (model, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS user_menu_permissions (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    menu_key VARCHAR(64) NOT NULL,
    granted_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, menu_key),
    CONSTRAINT user_menu_permissions_menu_key_check CHECK (menu_key IN ('conversation_logs.view_all'))
);
CREATE INDEX IF NOT EXISTS idx_user_menu_permissions_menu_key ON user_menu_permissions (menu_key, user_id);
