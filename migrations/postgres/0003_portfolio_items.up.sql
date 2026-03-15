CREATE TABLE IF NOT EXISTS portfolio_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR NOT NULL,
    description TEXT,
    image_url   VARCHAR,
    tags        TEXT[],          -- ['React', 'TypeScript', 'Node.js']
    order_index INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_portfolio_order ON portfolio_items(order_index);
