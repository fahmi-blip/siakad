-- 003_auth.sql
-- Inisialisasi tabel refresh_token untuk autentikasi

CREATE TABLE IF NOT EXISTS refresh_token (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     INTEGER      NOT NULL,
    token_hash  TEXT         NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ  NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_token_user_id_idx ON refresh_token (user_id);

-- Hubungkan FK ke users(id) jika tabel users sudah ada
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints 
            WHERE constraint_name = 'refresh_token_user_id_fkey' AND table_name = 'refresh_token'
        ) THEN
            ALTER TABLE refresh_token
                ADD CONSTRAINT refresh_token_user_id_fkey
                FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END IF;
END $$;