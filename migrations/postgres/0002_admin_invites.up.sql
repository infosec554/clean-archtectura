CREATE TABLE IF NOT EXISTS admin_invites (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      VARCHAR NOT NULL,
    token      VARCHAR NOT NULL UNIQUE,
    invited_by UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    used       BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_invites_token ON admin_invites(token);
CREATE INDEX IF NOT EXISTS idx_admin_invites_email ON admin_invites(email);
