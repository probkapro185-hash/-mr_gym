-- migrations/001_init.up.sql
-- CRM Gym — начальная схема БД

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    full_name     VARCHAR(255) NOT NULL,
    phone         VARCHAR(20)  NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL DEFAULT '',
    role          VARCHAR(20)  NOT NULL DEFAULT 'client'
                    CHECK (role IN ('client', 'manager', 'admin')),
    status        VARCHAR(20)  NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'approved', 'rejected')),
    balance       NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    visits        INTEGER        NOT NULL DEFAULT 0,
    notes         TEXT           NOT NULL DEFAULT '',
    trainer_id    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_role   ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_email  ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone  ON users(phone);

-- Комментарии к таблице
COMMENT ON TABLE users IS 'Пользователи системы: клиенты, менеджеры, администраторы';
COMMENT ON COLUMN users.status IS 'pending — заявка подана, approved — одобрена, rejected — отклонена';
COMMENT ON COLUMN users.trainer_id IS 'Закреплённый тренер (ссылается на другого пользователя с ролью manager/admin)';

-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS sessions (
    id          BIGSERIAL PRIMARY KEY,
    client_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trainer_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title       VARCHAR(255) NOT NULL DEFAULT 'Тренировка',
    description TEXT         NOT NULL DEFAULT '',
    start_time  TIMESTAMPTZ  NOT NULL,
    end_time    TIMESTAMPTZ  NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'scheduled'
                    CHECK (status IN ('scheduled', 'completed', 'cancelled')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_session_times CHECK (end_time > start_time)
);

CREATE INDEX IF NOT EXISTS idx_sessions_client  ON sessions(client_id, start_time);
CREATE INDEX IF NOT EXISTS idx_sessions_trainer ON sessions(trainer_id, start_time);
CREATE INDEX IF NOT EXISTS idx_sessions_start   ON sessions(start_time);

COMMENT ON TABLE sessions IS 'Расписание тренировок';

-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payments (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount         NUMERIC(12, 2) NOT NULL,
    service_name   VARCHAR(255)   NOT NULL DEFAULT '',
    operation_type VARCHAR(20)    NOT NULL DEFAULT 'deposit'
                        CHECK (operation_type IN ('deposit', 'charge', 'subscription')),
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_user       ON payments(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);

COMMENT ON TABLE payments IS 'Финансовые операции: пополнения, списания, абонементы';

-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS subscriptions (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255)   NOT NULL,
    visits_left INTEGER        NOT NULL DEFAULT 0,
    valid_until TIMESTAMPTZ    NOT NULL,
    price       NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id);

COMMENT ON TABLE subscriptions IS 'Абонементы клиентов';