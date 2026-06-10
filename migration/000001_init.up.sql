CREATE TABLE IF NOT EXISTS members (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(100) NOT NULL,
    avatar_color VARCHAR(7)   NOT NULL DEFAULT '#6366f1',
    is_active    BOOLEAN      NOT NULL DEFAULT true,
    sort_order   INT          NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_members_deleted_at ON members (deleted_at);

CREATE TABLE IF NOT EXISTS daily_queue (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_date DATE        NOT NULL,
    member_id  UUID        NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_daily_queue_queue_date ON daily_queue (queue_date);

CREATE TABLE IF NOT EXISTS queue_config (
    id             VARCHAR     PRIMARY KEY,
    last_member_id UUID,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS holidays (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    holiday_date DATE         NOT NULL,
    name         VARCHAR(200) NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_holidays_holiday_date UNIQUE (holiday_date)
);
