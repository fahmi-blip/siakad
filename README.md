# SIAKAD Mini - RESTful API Backend

Layanan akademik sederhana **SIAKAD Mini** (RESTful API) yang dibangun menggunakan **Go**, **Fiber v2**, dan **PostgreSQL** (via `pgx v5`).

---

## 📌 Teknologi & Stack Utama
- **Bahasa**: Go 1.26+
- **Framework Web**: Fiber v2
- **Database Driver**: PostgreSQL (`pgx/v5` dengan `pgxpool`)
- **Autentikasi**: JWT (Bearer Token)
- **Hasi Password**: `golang.org/x/crypto/bcrypt`
- **Validasi**: `go-playground/validator/v10`
- **Arsitektur**: Clean Architecture (`Model` -> `Repository` -> `Service` -> `Route/Handler`)

---

## 📁 Struktur Folder Utama
```text
.
├── app
│   ├── model          # Struct domain data & DTO request/response
│   ├── repository     # Query ke database PostgreSQL via pgx v5
│   └── service        # Business logic, aturan domain, & transaksi
├── cmd
│   └── seeder         # Executable script seeder data awal
├── config             # Konfigurasi aplikasi, env, logger, error handler
├── database           # Connection pool PostgreSQL
├── docs               # Postman collection & file pengujian HTTP
├── helper             # JWT manager, password hashing, validator, response helper
├── middleware         # Auth JWT, Role authorization, JSON, rate limiter
├── migrations         # File SQL migrasi database (up & down)
├── route              # Registrasi endpoint Fiber
├── main.go            # Entry point aplikasi
└── README.md          # Dokumentasi proyek
```

---

## 🚀 Setup & Panduan Menjalankan

### 1. Konfigurasi Environment Variable (`.env`)
Salin `.env.example` ke `.env` dan atur kredensial database & JWT:
```env
APP_PORT=3000
APP_ENV=development

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123456
DB_NAME=siakad
DB_SSLMODE=disable
DB_MAX_CONNS=10

JWT_SECRET=super_secret_jwt_key_that_is_at_least_32_chars_long
JWT_ISSUER=siakad-api
JWT_ACCESS_TTL_MINUTES=1440
```

### 2. Jalankan Migrasi Database
Jalankan file SQL migrasi di folder `migrations/` pada database PostgreSQL Anda:
```bash
psql -U postgres -d siakad -f migrations/008_siakad_schema.sql
```

### 3. Jalankan Seeder
Seeder bersifat **idempotent** (`ON CONFLICT DO NOTHING / UPDATE`) dan dapat dijalankan berulang tanpa duplikasi data:
```bash
go run ./cmd/seeder
```

### 4. Jalankan Server
```bash
go run main.go
```

---

## 👥 Akun Seed

### Admin
- **Email**: `admin@siakad.test`
- **Password**: `admin12345`

### Mahasiswa (20 Mahasiswa)
- **Password Awal**: Sama dengan **NIM** masing-masing.
- Contoh akun mahasiswa seed:
  - `mhs01@siakad.test` | NIM: `202401000001` | IPK: `3.75` (Maks 24 SKS)
  - `mhs03@siakad.test` | NIM: `202401000003` | IPK: `2.85` (Maks 21 SKS)
  - `mhs05@siakad.test` | NIM: `202401000005` | IPK: `2.30` (Maks 18 SKS)

---

## 📋 Daftar 10 Endpoint API (Prefix `/api/v1`)

| No | Method | Endpoint | Akses / Role | Deskripsi & Status Sukses |
|---|---|---|---|---|
| 1 | `POST` | `/api/v1/auth/login` | Publik | Login pengguna (Email & Password), mengembalikan Bearer JWT. (Status `200`) |
| 2 | `GET` | `/api/v1/auth/me` | Authenticated (Semua Role) | Mengembalikan profil pengguna login. Jika mahasiswa, menyertakan NIM, nama, prodi, angkatan. (Status `200`) |
| 3 | `GET` | `/api/v1/students` | Admin | Daftar mahasiswa dengan pagination, filter (`prodi`, `angkatan`), `search` (NIM/nama ILIKE), `sort` (`nama` atau `-ipk_terakhir`). (Status `200`) |
| 4 | `POST` | `/api/v1/students` | Admin | Menambah mahasiswa + akun user dalam 1 DB transaction. Password awal = NIM (hash bcrypt). (Status `201`) |
| 5 | `GET` | `/api/v1/students/{id}` | Admin / Mahasiswa (sendiri) | Detail mahasiswa + daftar mata kuliah yang diambil + `total_sks` + `batas_sks`. (Status `200`) |
| 6 | `PUT` | `/api/v1/students/{id}` | Admin | Memperbarui `nama`, `prodi`, `angkatan`, `ipk_terakhir` (NIM tidak boleh diubah). (Status `200`) |
| 7 | `DELETE` | `/api/v1/students/{id}` | Admin | Soft delete mahasiswa (`deleted_at`). Mahasiswa terhapus tidak bisa login. (Status `204`) |
| 8 | `GET` | `/api/v1/courses` | Authenticated (Semua Role) | Daftar mata kuliah + kalkulasi `terisi` & `sisa_kuota`. Filter: `semester`, `search`, `available=true`. (Status `200`) |
| 9 | `POST` | `/api/v1/enrollments` | Mahasiswa | Mengambil KRS (Mata Kuliah) dalam 1 DB transaction dengan `FOR UPDATE` row locking. (Status `201`) |
| 10 | `DELETE` | `/api/v1/enrollments/{id}` | Mahasiswa (milik sendiri) | Membatalkan pengambilan mata kuliah dari KRS. (Status `204`) |

---

## ⚙️ Aturan Bisnis (Business Rules)
1. **Batas SKS (MaxSKS)**:
   - IPK $\ge$ 3.00 $\rightarrow$ Maksimal **24 SKS**
   - IPK 2.50 – 2.99 $\rightarrow$ Maksimal **21 SKS**
   - IPK < 2.50 $\rightarrow$ Maksimal **18 SKS**
2. **Pengambilan Ganda**: Mahasiswa tidak boleh mengambil mata kuliah yang sama lebih dari sekali pada `tahun_akademik` yang sama.
3. **Kuota Mata Kuliah**: Mata kuliah dengan kuota penuh (`terisi >= kuota`) tidak boleh diambil. Dilengkapi `SELECT ... FOR UPDATE` row locking untuk mencegah race condition.
4. **Otorisasi Kepemilikan**: Mahasiswa hanya boleh melihat dan mengelola KRS/data milik sendiri.

---

## 🧪 Pengujian (Testing)

### 1. Jalankan Unit Test
```bash
go test -v ./app/service/...
```

### 2. Static Code Check & Build Verification
```bash
go vet ./...
go build ./...
```

### 3. File Postman & HTTP Request
- Postman Collection: `docs/SIAKAD_Mini_API.postman_collection.json`
- File HTTP Client: `docs/siakad_mini.http`
