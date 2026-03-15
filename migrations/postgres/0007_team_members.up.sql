CREATE TABLE IF NOT EXISTS team_members (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name  VARCHAR NOT NULL,
    last_name   VARCHAR NOT NULL,
    position    VARCHAR NOT NULL,   -- 'Bosh direktor', 'Lead dasturchi'
    avatar_url  VARCHAR,
    linkedin    VARCHAR,
    email       VARCHAR,
    order_index INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_team_members_order ON team_members(order_index);
