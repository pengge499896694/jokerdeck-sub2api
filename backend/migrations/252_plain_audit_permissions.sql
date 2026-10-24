CREATE TABLE IF NOT EXISTS plain_audit_log_permissions (
    viewer_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (viewer_user_id, target_user_id)
);
CREATE INDEX IF NOT EXISTS idx_plain_audit_permissions_target ON plain_audit_log_permissions(target_user_id);
