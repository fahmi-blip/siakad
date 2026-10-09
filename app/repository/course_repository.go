package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

type CourseRepository interface {
	FindAll(ctx context.Context, q model.CourseQuery) ([]model.Course, error)
	FindByID(ctx context.Context, id int) (model.Course, error)
	FindByIDTxForUpdate(ctx context.Context, tx pgx.Tx, id int) (model.Course, error)
}

type coursePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &coursePostgresRepository{pool: pool}
}

func (r *coursePostgresRepository) FindAll(ctx context.Context, q model.CourseQuery) ([]model.Course, error) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Semester != nil && *q.Semester > 0 {
		args = append(args, *q.Semester)
		where += fmt.Sprintf(" AND c.semester = $%d", len(args))
	}

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", len(args), len(args))
	}

	if q.Available {
		where += " AND (c.kuota - COALESCE(e.terisi, 0)) > 0"
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COALESCE(e.terisi, 0)::int AS terisi,
		       (c.kuota - COALESCE(e.terisi, 0))::int AS sisa_kuota
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) AS terisi
			FROM enrollments
			GROUP BY course_id
		) e ON e.course_id = c.id
		%s
		ORDER BY c.kode_mk ASC
	`, where)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar mata kuliah: %w", err)
	}
	defer rows.Close()

	courses := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota); err != nil {
			return nil, fmt.Errorf("membaca baris mata kuliah: %w", err)
		}
		courses = append(courses, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query mata kuliah: %w", err)
	}
	return courses, nil
}

func (r *coursePostgresRepository) FindByID(ctx context.Context, id int) (model.Course, error) {
	var c model.Course
	query := `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COALESCE(e.terisi, 0)::int AS terisi,
		       (c.kuota - COALESCE(e.terisi, 0))::int AS sisa_kuota
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) AS terisi
			FROM enrollments
			WHERE course_id = $1
			GROUP BY course_id
		) e ON e.course_id = c.id
		WHERE c.id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("mengambil mata kuliah: %w", err)
	}
	return c, nil
}

func (r *coursePostgresRepository) FindByIDTxForUpdate(ctx context.Context, tx pgx.Tx, id int) (model.Course, error) {
	var c model.Course
	query := `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COALESCE(e.terisi, 0)::int AS terisi,
		       (c.kuota - COALESCE(e.terisi, 0))::int AS sisa_kuota
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) AS terisi
			FROM enrollments
			WHERE course_id = $1
			GROUP BY course_id
		) e ON e.course_id = c.id
		WHERE c.id = $1
		FOR UPDATE OF c
	`
	err := tx.QueryRow(ctx, query, id).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("mengambil mata kuliah dengan row lock: %w", err)
	}
	return c, nil
}
