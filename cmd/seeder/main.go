package main

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"api-students/config"
	"api-students/database"
)

type UserSeed struct {
	Email    string
	Password string
	Role     string
}

type StudentSeed struct {
	UserEmail   string
	Nim         string
	Nama        string
	Prodi       string
	Angkatan    int
	IpkTerakhir float64
}

type CourseSeed struct {
	KodeMK   string
	NamaMK   string
	SKS      int
	Semester int
	Kuota    int
}

func main() {
	config.LoadEnv()

	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer pool.Close()

	fmt.Println("Memulai seeder SIAKAD Mini...")

	// 1. Seed Admin
	adminPassword := "admin12345"
	hashedAdminPass, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Gagal hash admin password: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, 'admin')
		ON CONFLICT (email) DO NOTHING
	`, "admin@siakad.test", string(hashedAdminPass))
	if err != nil {
		log.Fatalf("Gagal seed admin: %v", err)
	}
	fmt.Println("✔ Admin seeded")

	// 2. Seed 20 Mahasiswa (Users + Students)
	studentsData := []StudentSeed{
		{"mhs01@siakad.test", "202401000001", "Budi Santoso", "Teknik Informatika", 2024, 3.75},
		{"mhs02@siakad.test", "202401000002", "Siti Rahma", "Sistem Informasi", 2024, 3.40},
		{"mhs03@siakad.test", "202401000003", "Ahmad Fauzi", "Teknik Informatika", 2023, 2.85},
		{"mhs04@siakad.test", "202401000004", "Dewi Lestari", "Teknologi Informasi", 2023, 2.60},
		{"mhs05@siakad.test", "202401000005", "Rian Hidayat", "Teknik Informatika", 2022, 2.30},
		{"mhs06@siakad.test", "202401000006", "Maya Putri", "Sistem Informasi", 2022, 2.10},
		{"mhs07@siakad.test", "202401000007", "Eko Prasetyo", "Teknologi Informasi", 2024, 3.90},
		{"mhs08@siakad.test", "202401000008", "Fitriani", "Teknik Informatika", 2024, 3.20},
		{"mhs09@siakad.test", "202401000009", "Gita Gutawa", "Sistem Informasi", 2023, 2.95},
		{"mhs10@siakad.test", "202401000010", "Hendra Gunawan", "Teknologi Informasi", 2023, 2.70},
		{"mhs11@siakad.test", "202401000011", "Indah Permata", "Teknik Informatika", 2022, 2.45},
		{"mhs12@siakad.test", "202401000012", "Joko Widodo", "Sistem Informasi", 2022, 1.95},
		{"mhs13@siakad.test", "202401000013", "Kiki Amelia", "Teknologi Informasi", 2024, 3.55},
		{"mhs14@siakad.test", "202401000014", "Lukman Hakim", "Teknik Informatika", 2024, 3.10},
		{"mhs15@siakad.test", "202401000015", "Nia Ramadhani", "Sistem Informasi", 2023, 2.80},
		{"mhs16@siakad.test", "202401000016", "Oscar Lawalata", "Teknologi Informasi", 2023, 2.55},
		{"mhs17@siakad.test", "202401000017", "Putri Marino", "Teknik Informatika", 2022, 2.25},
		{"mhs18@siakad.test", "202401000018", "Qory Sandioriva", "Sistem Informasi", 2022, 3.85},
		{"mhs19@siakad.test", "202401000019", "Raffi Ahmad", "Teknologi Informasi", 2024, 2.75},
		{"mhs20@siakad.test", "202401000020", "Sule Priakitiew", "Teknik Informatika", 2024, 2.40},
	}

	for _, s := range studentsData {
		hashedNimPass, err := bcrypt.GenerateFromPassword([]byte(s.Nim), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Gagal hash password NIM %s: %v", s.Nim, err)
		}

		var userID int
		err = pool.QueryRow(ctx, `
			INSERT INTO users (email, password, role)
			VALUES ($1, $2, 'mahasiswa')
			ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
			RETURNING id
		`, s.UserEmail, string(hashedNimPass)).Scan(&userID)
		if err != nil {
			log.Fatalf("Gagal insert user %s: %v", s.UserEmail, err)
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (nim) DO UPDATE SET
				nama = EXCLUDED.nama,
				prodi = EXCLUDED.prodi,
				angkatan = EXCLUDED.angkatan,
				ipk_terakhir = EXCLUDED.ipk_terakhir
		`, userID, s.Nim, s.Nama, s.Prodi, s.Angkatan, s.IpkTerakhir)
		if err != nil {
			log.Fatalf("Gagal insert student NIM %s: %v", s.Nim, err)
		}
	}
	fmt.Println("✔ 20 Mahasiswa seeded")

	// 3. Seed 10 Mata Kuliah
	coursesData := []CourseSeed{
		{"IF101", "Algoritma dan Pemrograman", 3, 1, 30},
		{"IF102", "Struktur Data", 4, 2, 25},
		{"IF103", "Pemrograman Web", 3, 3, 20},
		{"IF104", "Basis Data", 3, 2, 35},
		{"IF105", "Jaringan Komputer", 3, 4, 30},
		{"IF106", "Kecerdasan Buatan", 3, 5, 15},
		{"IF107", "Sistem Operasi", 3, 3, 25},
		{"IF108", "Rekayasa Perangkat Lunak", 3, 4, 20},
		{"IF109", "Seminar Hasil", 2, 7, 2},  // Kuota kecil (2) untuk uji kuota penuh
		{"IF110", "Skripsi", 6, 8, 1},       // Kuota sangat kecil (1) untuk uji kuota penuh
	}

	for _, c := range coursesData {
		_, err := pool.Exec(ctx, `
			INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (kode_mk) DO UPDATE SET
				nama_mk = EXCLUDED.nama_mk,
				sks = EXCLUDED.sks,
				semester = EXCLUDED.semester,
				kuota = EXCLUDED.kuota
		`, c.KodeMK, c.NamaMK, c.SKS, c.Semester, c.Kuota)
		if err != nil {
			log.Fatalf("Gagal seed course %s: %v", c.KodeMK, err)
		}
	}
	fmt.Println("✔ 10 Mata kuliah seeded")

	fmt.Println("Seeder selesai dengan sukses!")
}
