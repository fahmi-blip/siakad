-- 002_create_prestasi.sql
-- Inisialisasi tabel prestasi mahasiswa (relasi 1-to-many dengan students)

CREATE TABLE IF NOT EXISTS prestasi (
    id                  SERIAL          PRIMARY KEY,
    student_id          INT             NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    name_prestation     VARCHAR(100)    NOT NULL,
    juara               VARCHAR(50)     NOT NULL
);

CREATE INDEX IF NOT EXISTS prestation_student_id_idx ON prestasi (student_id);