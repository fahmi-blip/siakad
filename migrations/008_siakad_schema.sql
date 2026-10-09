-- 008_siakad_schema.sql
-- Skema Database Lengkap SIAKAD Mini (Satu Kesatuan Terhubung)
-- Urutan eksekusi mengikuti dependensi relasional foreign key secara tepat.

-- ============================================================================
-- 1. ROLES & PERMISSIONS (RBAC)
-- ============================================================================
CREATE TABLE IF NOT EXISTS roles (
    name            VARCHAR(20)     PRIMARY KEY,
    description     VARCHAR(150)    NOT NULL,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description)
SELECT 'admin', 'Akses penuh terhadap seluruh data dan pengaturan'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'admin');

INSERT INTO roles (name, description)
SELECT 'mahasiswa', 'Akses mahasiswa untuk melihat data sendiri dan mengelola KRS'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'mahasiswa');

INSERT INTO roles (name, description)
SELECT 'staff', 'Boleh melihat data seluruh user, tetapi tidak boleh mengubah'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'staff');

CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50)     PRIMARY KEY,
    description VARCHAR(150)    NOT NULL
);

INSERT INTO permissions (name, description)
SELECT 'student:list', 'Melihat daftar seluruh student'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:list');
INSERT INTO permissions (name, description)
SELECT 'student:read:any', 'Melihat data student mana pun'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:read:any');
INSERT INTO permissions (name, description)
SELECT 'student:read:self', 'Melihat data profil student sendiri'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:read:self');
INSERT INTO permissions (name, description)
SELECT 'student:create', 'Membuat data student baru'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:create');
INSERT INTO permissions (name, description)
SELECT 'student:update:any', 'Mengubah data student mana pun'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:update:any');
INSERT INTO permissions (name, description)
SELECT 'student:delete', 'Menghapus (soft delete) student'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'student:delete');
INSERT INTO permissions (name, description)
SELECT 'role:assign', 'Mengubah role milik pengguna lain'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'role:assign');
INSERT INTO permissions (name, description)
SELECT 'course:list', 'Melihat daftar mata kuliah'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'course:list');
INSERT INTO permissions (name, description)
SELECT 'enrollment:create', 'Mengambil KRS mata kuliah'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'enrollment:create');
INSERT INTO permissions (name, description)
SELECT 'enrollment:delete', 'Membatalkan KRS mata kuliah'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'enrollment:delete');

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20)     NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50)     NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

-- Admin Permissions
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:list' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:list');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:read:any' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:read:any');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:create' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:create');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:update:any' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:update:any');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:delete' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:delete');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'role:assign' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'role:assign');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'course:list' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'course:list');

-- Staff Permissions
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'student:list' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'student:list');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'student:read:any' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'student:read:any');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'student:create' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'student:create');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'course:list' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'course:list');

-- Mahasiswa Permissions
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'mahasiswa', 'student:read:self' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'mahasiswa' AND permission_name = 'student:read:self');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'mahasiswa', 'course:list' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'mahasiswa' AND permission_name = 'course:list');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'mahasiswa', 'enrollment:create' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'mahasiswa' AND permission_name = 'enrollment:create');
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'mahasiswa', 'enrollment:delete' WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'mahasiswa' AND permission_name = 'enrollment:delete');

-- ============================================================================
-- 2. USERS
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id          SERIAL          PRIMARY KEY,
    email       VARCHAR(255)    NOT NULL UNIQUE,
    password    VARCHAR(255)    NOT NULL,
    role        VARCHAR(20)     NOT NULL DEFAULT 'mahasiswa' REFERENCES roles(name) ON UPDATE CASCADE ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_idx ON users (LOWER(email));
CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);
CREATE INDEX IF NOT EXISTS users_created_at_id_desc_idx ON users (created_at DESC, id DESC);

-- ============================================================================
-- 3. REFRESH TOKEN (AUTH)
-- ============================================================================
CREATE TABLE IF NOT EXISTS refresh_token (
    id          BIGSERIAL       PRIMARY KEY,
    user_id     INTEGER         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT            NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ     NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_token_user_id_idx ON refresh_token (user_id);

-- ============================================================================
-- 4. STUDENTS
-- ============================================================================
CREATE TABLE IF NOT EXISTS students (
    id           SERIAL          PRIMARY KEY,
    user_id      INT             UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nim          VARCHAR(12)     UNIQUE NOT NULL,
    nama         VARCHAR(100)    NOT NULL,
    prodi        VARCHAR(100)    NOT NULL,
    angkatan     INT             NOT NULL,
    ipk_terakhir NUMERIC(3,2)    NOT NULL DEFAULT 0.00 CHECK (ipk_terakhir >= 0.00 AND ipk_terakhir <= 4.00),
    deleted_at   TIMESTAMPTZ     NULL,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS students_prodi_idx ON students (prodi);
CREATE INDEX IF NOT EXISTS students_angkatan_idx ON students (angkatan);
CREATE INDEX IF NOT EXISTS students_user_id_idx ON students (user_id);
CREATE INDEX IF NOT EXISTS students_deleted_at_idx ON students (deleted_at);
CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx ON students (created_at DESC, id DESC);

-- ============================================================================
-- 5. PRESTASI
-- ============================================================================
CREATE TABLE IF NOT EXISTS prestasi (
    id                  SERIAL          PRIMARY KEY,
    student_id          INT             NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    name_prestation     VARCHAR(100)    NOT NULL,
    juara               VARCHAR(50)     NOT NULL
);

CREATE INDEX IF NOT EXISTS prestation_student_id_idx ON prestasi (student_id);

-- ============================================================================
-- 6. COURSES (MATA KULIAH)
-- ============================================================================
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

-- ============================================================================
-- 7. ENROLLMENTS (KRS)
-- ============================================================================
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
