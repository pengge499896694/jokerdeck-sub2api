CREATE TABLE IF NOT EXISTS recharge_terms_acceptances (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version VARCHAR(32) NOT NULL,
    not_mainland_subject BOOLEAN NOT NULL CHECK (not_mainland_subject),
    outside_mainland BOOLEAN NOT NULL CHECK (outside_mainland),
    no_invoice BOOLEAN NOT NULL CHECK (no_invoice),
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, version)
);
