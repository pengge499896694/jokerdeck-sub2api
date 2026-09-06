CREATE TABLE IF NOT EXISTS xianyu_recharge_orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    xianyu_order_no VARCHAR(128) NOT NULL UNIQUE,
    paid_amount_cny DECIMAL(20, 2) NOT NULL CHECK (paid_amount_cny > 0),
    note TEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'credited', 'rejected')),
    credited_amount_usd DECIMAL(20, 8) NULL,
    admin_note TEXT NULL,
    reviewed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS xianyu_recharge_orders_user_id_idx
    ON xianyu_recharge_orders (user_id);

CREATE INDEX IF NOT EXISTS xianyu_recharge_orders_status_created_at_idx
    ON xianyu_recharge_orders (status, created_at DESC);

