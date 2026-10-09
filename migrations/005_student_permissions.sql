-- 005_student_permissions.sql
-- Role-Based Access Control (RBAC): Tabel roles, permissions, role_permissions,
-- serta relasi foreign key dari users ke roles.

-- 1. Tabel Roles
CREATE TABLE IF NOT EXISTS roles (
    name            VARCHAR(20)        PRIMARY KEY,
    description     VARCHAR(150)       NOT NULL,
    created_at      TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

-- Seed data role (admin, mahasiswa, staff, dan user sebagai alias)
INSERT INTO roles (name, description)
SELECT 'admin', 'Akses penuh terhadap seluruh data dan pengaturan'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'admin');

INSERT INTO roles (name, description)
SELECT 'mahasiswa', 'Akses mahasiswa untuk melihat data sendiri dan mengelola KRS'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'mahasiswa');

INSERT INTO roles (name, description)
SELECT 'staff', 'Boleh melihat data seluruh user, tetapi tidak boleh mengubah'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'staff');

INSERT INTO roles (name, description)
SELECT 'user', 'Pengguna umum aplikasi'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'user');

-- 2. Tabel Permissions
CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50)     PRIMARY KEY,
    description VARCHAR(150)    NOT NULL
);

-- Seed data permissions
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

-- 3. Tabel Role Permissions (Many-to-Many)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20)     NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50)     NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

-- Seed relasi role_permissions untuk 'admin'
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:list'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:list');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:read:any'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:read:any');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:create'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:create');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:update:any'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:update:any');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'student:delete'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'student:delete');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'role:assign'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'role:assign');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', 'course:list'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'admin' AND permission_name = 'course:list');

-- Seed relasi role_permissions untuk 'staff'
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'student:list'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'student:list');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'student:read:any'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'student:read:any');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'student:create'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'student:create');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'staff', 'course:list'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'staff' AND permission_name = 'course:list');

-- Seed relasi role_permissions untuk 'mahasiswa'
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'mahasiswa', 'student:read:self'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'mahasiswa' AND permission_name = 'student:read:self');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'mahasiswa', 'course:list'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'mahasiswa' AND permission_name = 'course:list');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'mahasiswa', 'enrollment:create'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'mahasiswa' AND permission_name = 'enrollment:create');

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'mahasiswa', 'enrollment:delete'
WHERE NOT EXISTS (SELECT 1 FROM role_permissions WHERE role_name = 'mahasiswa' AND permission_name = 'enrollment:delete');

-- 4. Hubungkan tabel users ke roles
-- Hapus check constraint lama jika ada agar tidak konflik dengan tabel roles
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

-- Pastikan semua nilai role di users valid terhadap tabel roles
UPDATE users SET role = 'mahasiswa' WHERE role NOT IN (SELECT name FROM roles);

-- Pasang foreign key dari users(role) ke roles(name) secara aman
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey
    FOREIGN KEY (role) REFERENCES roles(name)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);
