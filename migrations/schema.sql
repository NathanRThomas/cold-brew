\c coldbrew;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id              UUID NOT NULL PRIMARY KEY,
    email           CITEXT NOT NULL UNIQUE,
    token           TEXT NOT NULL UNIQUE,
    attr            JSONB NOT NULL DEFAULT '{}',
    mask            INT NOT NULL DEFAULT 0,
    disabled        TIMESTAMPTZ,
    created         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_disabled ON users (disabled);

CREATE TABLE mailmen (
    id              UUID NOT NULL PRIMARY KEY,
    attr            JSONB NOT NULL DEFAULT '{}',
    mask            INT NOT NULL DEFAULT 0,
    created         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE templates (
    id              UUID NOT NULL PRIMARY KEY,
    body_html       TEXT NOT NULL,
    body_text       TEXT NOT NULL,
    subject         TEXT NOT NULL,
    preview_text    TEXT NOT NULL,
    attr            JSONB NOT NULL DEFAULT '{}',
    mask            INT NOT NULL DEFAULT 0,
    created         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE emails (
    id              UUID NOT NULL PRIMARY KEY,
    target_time     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_time       TIMESTAMPTZ,
    mailman         UUID NOT NULL REFERENCES mailmen (id) ON DELETE CASCADE,
    template        UUID NOT NULL REFERENCES templates (id) ON DELETE CASCADE,
    "user"          UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status          TEXT NOT NULL,
    message_id      TEXT NOT NULL,
    mask            INT NOT NULL DEFAULT 0,
    created         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_emails_message_id ON emails (message_id);
CREATE INDEX idx_emails_target_time ON emails (target_time);
CREATE INDEX idx_emails_sent_time ON emails (sent_time);
