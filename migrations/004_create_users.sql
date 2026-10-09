-- 004_create_users.sql
-- Inisialisasi tabel users untuk manajemen pengguna dan autentikasi

CREATE TABLE IF NOT EXISTS users (
    id          SERIAL          PRIMARY KEY,
    email       VARCHAR(255)    NOT NULL UNIQUE,
    password    VARCHAR(255)    NOT NULL,
    role        VARCHAR(20)     NOT NULL DEFAULT 'mahasiswa',
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Kolom idempotent jika tabel sudah ada sebelumnya
ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS password VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'mahasiswa';

-- Pastikan indeks unik case-insensitive untuk email dan indeks role
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_idx ON users (LOWER(email));
CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);

-- Hubungkan FK dari students.user_id -> users(id) jika tabel students sudah ada
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'students') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints 
            WHERE constraint_name = 'students_user_id_fkey' AND table_name = 'students'
        ) THEN
            ALTER TABLE students
                ADD CONSTRAINT students_user_id_fkey
                FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END IF;
END $$;

-- Hubungkan FK dari refresh_token.user_id -> users(id) jika tabel refresh_token sudah ada
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'refresh_token') THEN
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

-- Hubungkan users.role ke roles(name) jika tabel roles sudah ada
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints 
            WHERE constraint_name = 'users_role_fkey' AND table_name = 'users'
        ) THEN
            ALTER TABLE users
                ADD CONSTRAINT users_role_fkey
                FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE ON DELETE RESTRICT;
        END IF;
    END IF;
END $$;