CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'requester'
                  CHECK (role IN ('requester', 'agent', 'admin')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uq
    ON users (lower(email));

CREATE TABLE IF NOT EXISTS tickets (
    id           BIGSERIAL PRIMARY KEY,
    title        VARCHAR(150) NOT NULL,
    description  TEXT         NOT NULL,
    category     VARCHAR(20)  NOT NULL
                 CHECK (category IN ('hardware', 'software', 'network', 'access', 'other')),
    priority     VARCHAR(20)  NOT NULL DEFAULT 'medium'
                 CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    status       VARCHAR(20)  NOT NULL DEFAULT 'open'
                 CHECK (status IN ('open', 'in_progress', 'resolved', 'closed')),
    requester_id BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assignee_id  BIGINT       REFERENCES users(id) ON DELETE RESTRICT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS tickets_requester_idx ON tickets (requester_id);
CREATE INDEX IF NOT EXISTS tickets_assignee_idx  ON tickets (assignee_id);
CREATE INDEX IF NOT EXISTS tickets_status_idx    ON tickets (status);
CREATE INDEX IF NOT EXISTS tickets_priority_idx  ON tickets (priority);
CREATE INDEX IF NOT EXISTS tickets_created_idx   ON tickets (created_at DESC);
CREATE INDEX IF NOT EXISTS tickets_updated_idx   ON tickets (updated_at DESC);

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS tickets_title_trgm_idx
    ON tickets USING gin (title gin_trgm_ops);

CREATE TABLE IF NOT EXISTS comments (
    id          BIGSERIAL PRIMARY KEY,
    ticket_id   BIGINT      NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author_id   BIGINT      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    body        TEXT        NOT NULL,
    is_internal BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS comments_ticket_idx ON comments (ticket_id, created_at);
