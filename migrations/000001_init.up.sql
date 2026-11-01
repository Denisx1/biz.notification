CREATE TABLE sms_tasks (
    id         BIGSERIAL PRIMARY KEY,
    phone      TEXT        NOT NULL,
    text       TEXT        NOT NULL,
    ip         TEXT        NOT NULL DEFAULT '',
    status     TEXT        NOT NULL DEFAULT 'pending'
               CHECK (status IN ('pending', 'sent', 'failed', 'expired')),
    error      TEXT        NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at    TIMESTAMPTZ
);

CREATE INDEX sms_tasks_status_created_idx ON sms_tasks (status, created_at);
CREATE INDEX sms_tasks_phone_created_idx  ON sms_tasks (phone, created_at);