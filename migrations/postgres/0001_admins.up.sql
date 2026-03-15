CREATE TABLE IF NOT EXISTS admins (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      VARCHAR NOT NULL UNIQUE,
    password   VARCHAR NOT NULL,
    role       VARCHAR NOT NULL DEFAULT 'admin' CHECK (role IN ('superadmin', 'admin')),
    invited_by UUID REFERENCES admins(id) ON DELETE SET NULL,
    is_active  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admins_email ON admins(email);

-- Default superadmin (parolni .env dan o'zgartiring)
-- password: "admin123" (bcrypt hash — production'da o'zgartiring)
INSERT INTO admins (email, password, role, is_active)
VALUES ('admin@example.com', 'admin123', 'superadmin', TRUE)
ON CONFLICT (email) DO NOTHING;
