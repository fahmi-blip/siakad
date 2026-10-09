CREATE TABLE IF NOT EXISTS enrollments (
    id             SERIAL          PRIMARY KEY,
    student_id     INT             NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id      INT             NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    tahun_akademik VARCHAR(30)     NOT NULL,
    created_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    CONSTRAINT enrollments_student_course_tahun_unique UNIQUE (student_id, course_id, tahun_akademik)
);

CREATE INDEX IF NOT EXISTS enrollments_course_id_idx ON enrollments (course_id);
CREATE INDEX IF NOT EXISTS enrollments_student_id_idx ON enrollments (student_id);
CREATE INDEX IF NOT EXISTS enrollments_tahun_akademik_idx ON enrollments (tahun_akademik);
