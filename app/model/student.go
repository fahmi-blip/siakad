package model

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Student struct {
	ID          int                `json:"id"`
	UserID      int                `json:"user_id"`
	Nim         string             `json:"nim"`
	Nama        string             `json:"nama"`
	Prodi       string             `json:"prodi"`
	Angkatan    int                `json:"angkatan"`
	IpkTerakhir float64            `json:"ipk_terakhir"`
	DeletedAt   pgtype.Timestamptz `json:"-"`
	CreatedAt   time.Time          `json:"created_at,omitempty"`
}

type StudentDetailResponse struct {
	ID          int          `json:"id"`
	UserID      int          `json:"user_id"`
	Nim         string       `json:"nim"`
	Nama        string       `json:"nama"`
	Prodi       string       `json:"prodi"`
	Angkatan    int          `json:"angkatan"`
	IpkTerakhir float64      `json:"ipk_terakhir"`
	MataKuliah  []CourseItem `json:"mata_kuliah"`
	TotalSKS    int          `json:"total_sks"`
	BatasSKS    int          `json:"batas_sks"`
	CreatedAt   time.Time    `json:"created_at,omitempty"`
}

type CourseItem struct {
	ID            int    `json:"id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	Semester      int    `json:"semester"`
	TahunAkademik string `json:"tahun_akademik,omitempty"`
}

type CreateStudentRequest struct {
	Nim         string   `json:"nim" validate:"required,nim"`
	Nama        string   `json:"nama" validate:"required,min=2,max=100"`
	Email       string   `json:"email" validate:"required,email"`
	Prodi       string   `json:"prodi" validate:"required,min=2,max=100"`
	Angkatan    int      `json:"angkatan" validate:"required"`
	IpkTerakhir *float64 `json:"ipk_terakhir" validate:"omitempty,gte=0,lte=4"`
}

type UpdateStudentRequest struct {
	Nama        string   `json:"nama" validate:"required,min=2,max=100"`
	Prodi       string   `json:"prodi" validate:"required,min=2,max=100"`
	Angkatan    int      `json:"angkatan" validate:"required"`
	IpkTerakhir *float64 `json:"ipk_terakhir" validate:"omitempty,gte=0,lte=4"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

type StudentListQuery struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan int
	Search   string
	Sort     string
}

func (q StudentListQuery) Offset() int {
	return (q.Page - 1) * q.PerPage
}

type Prestation struct {
	ID             int    `json:"id"`
	StudentID      int    `json:"student_id"`
	NamePrestation string `json:"name_prestation"`
	Juara          string `json:"juara"`
}

type StudentWithPrestation struct {
	Student
	Prestasi []Prestation `json:"prestasi"`
}
