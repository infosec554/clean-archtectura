CREATE TABLE IF NOT EXISTS testimonials (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_name  VARCHAR NOT NULL,
    author_role  VARCHAR,           -- 'Bosh texnologiya direktori, Meridian Labs'
    avatar_url   VARCHAR,
    content      TEXT NOT NULL,
    rating       SMALLINT NOT NULL DEFAULT 5 CHECK (rating BETWEEN 1 AND 5),
    order_index  INT NOT NULL DEFAULT 0,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMP DEFAULT NOW(),
    updated_at   TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_testimonials_order ON testimonials(order_index);
