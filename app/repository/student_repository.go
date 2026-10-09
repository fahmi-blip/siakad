package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah terdaftar")
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.StudentListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindByUserID(ctx context.Context, userID int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	CreateTx(ctx context.Context, tx pgx.Tx, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
	FindByIDWithCourses(ctx context.Context, id int) (model.StudentDetailResponse, error)
	FindByIDWithPrestasi(ctx context.Context, id int) (model.StudentWithPrestation, error)
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

var allowedSortColumns = map[string]string{
	"nama":          "nama ASC",
	"-ipk_terakhir": "ipk_terakhir DESC",
}

func (r *studentPostgresRepository) FindAll(
	ctx context.Context, q model.StudentListQuery,
) ([]model.Student, int, error) {
	where := " WHERE deleted_at IS NULL"
	args := []any{}

	if q.Prodi != "" {
		args = append(args, q.Prodi)
		where += fmt.Sprintf(" AND prodi = $%d", len(args))
	}
	if q.Angkatan > 0 {
		args = append(args, q.Angkatan)
		where += fmt.Sprintf(" AND angkatan = $%d", len(args))
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (nim ILIKE $%d OR nama ILIKE $%d)", len(args), len(args))
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM students" + where
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung total student: %w", err)
	}

	orderBy := "ORDER BY id ASC"
	if sorted, ok := allowedSortColumns[q.Sort]; ok {
		orderBy = "ORDER BY " + sorted
	}

	sqlText := fmt.Sprintf(
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at
		 FROM students%s
		 %s
		 LIMIT $%d OFFSET $%d`,
		where, orderBy, len(args)+1, len(args)+2,
	)
	args = append(args, q.PerPage, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar student: %w", err)
	}
	defer rows.Close()

	list := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.Nim, &s.Nama, &s.Prodi, &s.Angkatan, &s.IpkTerakhir, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris student: %w", err)
		}
		list = append(list, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query: %w", err)
	}
	return list, total, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	var deletedAt *time.Time

	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at
		 FROM students WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&s.ID, &s.UserID, &s.Nim, &s.Nama, &s.Prodi, &s.Angkatan, &s.IpkTerakhir, &deletedAt, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) FindByUserID(ctx context.Context, userID int) (model.Student, error) {
	var s model.Student
	var deletedAt *time.Time

	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at
		 FROM students WHERE user_id = $1 AND deleted_at IS NULL`, userID,
	).Scan(&s.ID, &s.UserID, &s.Nim, &s.Nama, &s.Prodi, &s.Angkatan, &s.IpkTerakhir, &deletedAt, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student dari user_id: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Create(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		s.UserID, s.Nim, s.Nama, s.Prodi, s.Angkatan, s.IpkTerakhir,
	).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan student: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) CreateTx(ctx context.Context, tx pgx.Tx, s model.Student) (model.Student, error) {
	err := tx.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		s.UserID, s.Nim, s.Nama, s.Prodi, s.Angkatan, s.IpkTerakhir,
	).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan student dalam transaksi: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE students
		 SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4
		 WHERE id = $5 AND deleted_at IS NULL
		 RETURNING user_id, nim, created_at`,
		s.Nama, s.Prodi, s.Angkatan, s.IpkTerakhir, s.ID,
	).Scan(&s.UserID, &s.Nim, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("memperbarui student: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) SoftDelete(ctx context.Context, id int) error {
	cmd, err := r.pool.Exec(ctx,
		`UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("menghapus student: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *studentPostgresRepository) FindByIDWithCourses(ctx context.Context, id int) (model.StudentDetailResponse, error) {
	student, err := r.FindByID(ctx, id)
	if err != nil {
		return model.StudentDetailResponse{}, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1
		 ORDER BY c.id ASC`, id,
	)
	if err != nil {
		return model.StudentDetailResponse{}, fmt.Errorf("mengambil mata kuliah student: %w", err)
	}
	defer rows.Close()

	courses := []model.CourseItem{}
	totalSKS := 0
	for rows.Next() {
		var item model.CourseItem
		if err := rows.Scan(&item.ID, &item.KodeMK, &item.NamaMK, &item.SKS, &item.Semester, &item.TahunAkademik); err != nil {
			return model.StudentDetailResponse{}, fmt.Errorf("membaca mata kuliah student: %w", err)
		}
		courses = append(courses, item)
		totalSKS += item.SKS
	}
	if err := rows.Err(); err != nil {
		return model.StudentDetailResponse{}, fmt.Errorf("membaca hasil query mata kuliah: %w", err)
	}

	return model.StudentDetailResponse{
		ID:          student.ID,
		UserID:      student.UserID,
		Nim:         student.Nim,
		Nama:        student.Nama,
		Prodi:       student.Prodi,
		Angkatan:    student.Angkatan,
		IpkTerakhir: student.IpkTerakhir,
		MataKuliah:  courses,
		TotalSKS:    totalSKS,
		CreatedAt:   student.CreatedAt,
	}, nil
}

func (r *studentPostgresRepository) FindByIDWithPrestasi(ctx context.Context, id int) (model.StudentWithPrestation, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.user_id, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir, s.created_at,
				p.id, p.student_id, p.name_prestation, p.juara
		 FROM students s
		 LEFT JOIN prestasi p ON p.student_id = s.id
		 WHERE s.id = $1 AND s.deleted_at IS NULL
		 ORDER BY p.id`, id,
	)
	if err != nil {
		return model.StudentWithPrestation{}, fmt.Errorf("mengambil student dengan prestasi: %w", err)
	}
	defer rows.Close()

	var hasil model.StudentWithPrestation
	ditemukan := false
	for rows.Next() {
		var s model.Student
		var pID, pStudentID sql.NullInt64
		var pNamePrestation, pJuara sql.NullString

		if err := rows.Scan(
			&s.ID, &s.UserID, &s.Nim, &s.Nama, &s.Prodi, &s.Angkatan, &s.IpkTerakhir, &s.CreatedAt,
			&pID, &pStudentID, &pNamePrestation, &pJuara,
		); err != nil {
			return model.StudentWithPrestation{}, fmt.Errorf("membaca baris prestasi: %w", err)
		}
		if !ditemukan {
			hasil.Student = s
			hasil.Prestasi = []model.Prestation{}
			ditemukan = true
		}
		if pID.Valid {
			hasil.Prestasi = append(hasil.Prestasi, model.Prestation{
				ID:             int(pID.Int64),
				StudentID:      int(pStudentID.Int64),
				NamePrestation: pNamePrestation.String,
				Juara:          pJuara.String,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return model.StudentWithPrestation{}, fmt.Errorf("membaca hasil query: %w", err)
	}
	if !ditemukan {
		return model.StudentWithPrestation{}, ErrNotFound
	}
	return hasil, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
