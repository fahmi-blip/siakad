-- 001_create_students.sql
-- Inisialisasi tabel students

CREATE TABLE IF NOT EXISTS students (
    id           SERIAL          PRIMARY KEY,
    user_id      INT             UNIQUE,
    nim          VARCHAR(12)     UNIQUE NOT NULL,
    nama         VARCHAR(100)    NOT NULL,
    prodi        VARCHAR(100)    NOT NULL,
    angkatan     INT             NOT NULL,
    ipk_terakhir NUMERIC(3,2)    NOT NULL DEFAULT 0.00 CHECK (ipk_terakhir >= 0.00 AND ipk_terakhir <= 4.00),
    deleted_at   TIMESTAMPTZ     NULL,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Kolom idempotent jika tabel sudah ada dari versi sebelumnya
ALTER TABLE students ADD COLUMN IF NOT EXISTS user_id INT UNIQUE;
ALTER TABLE students ADD COLUMN IF NOT EXISTS nama VARCHAR(100);
ALTER TABLE students ADD COLUMN IF NOT EXISTS prodi VARCHAR(100);
ALTER TABLE students ADD COLUMN IF NOT EXISTS angkatan INT;
ALTER TABLE students ADD COLUMN IF NOT EXISTS ipk_terakhir NUMERIC(3,2) DEFAULT 0.00;
ALTER TABLE students ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

-- Index pendukung pencarian dan filter
CREATE INDEX IF NOT EXISTS students_prodi_idx ON students (prodi);
CREATE INDEX IF NOT EXISTS students_angkatan_idx ON students (angkatan);
CREATE INDEX IF NOT EXISTS students_user_id_idx ON students (user_id);
CREATE INDEX IF NOT EXISTS students_deleted_at_idx ON students (deleted_at);

-- Hubungkan FK ke users(id) jika tabel users sudah ada
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
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