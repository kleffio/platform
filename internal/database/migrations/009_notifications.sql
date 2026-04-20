-- System-wide notification inbox.
-- Each row is a notification for a specific user (identified by their IDP subject / user_id).
-- Notifications are soft-readable: read_at is NULL until the user marks it read.

CREATE TABLE IF NOT EXISTS notifications (
    id         TEXT        PRIMARY KEY,
    user_id    TEXT        NOT NULL,
    type       TEXT        NOT NULL,
    title      TEXT        NOT NULL,
    body       TEXT        NOT NULL DEFAULT '',
    data       JSONB       NOT NULL DEFAULT '{}',
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Most queries filter by user_id and sort by created_at desc.
CREATE INDEX IF NOT EXISTS idx_notifications_user_id
    ON notifications(user_id, created_at DESC);

-- Fast unread count per user.
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications(user_id)
    WHERE read_at IS NULL;
