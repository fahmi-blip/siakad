CREATE TABLE IF NOT EXISTS courses (
    id          SERIAL          PRIMARY KEY,
    kode_mk     VARCHAR(20)     UNIQUE NOT NULL,
    nama_mk     VARCHAR(100)    NOT NULL,
    sks         INT             NOT NULL CHECK (sks > 0),
    semester    INT             NOT NULL CHECK (semester > 0),
    kuota       INT             NOT NULL CHECK (kuota >= 0),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS courses_kode_mk_key ON courses (kode_mk);