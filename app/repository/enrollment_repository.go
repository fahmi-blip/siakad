package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

type EnrollmentRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, e model.Enrollment) (model.Enrollment, error)
	CountByCourseTx(ctx context.Context, tx pgx.Tx, courseID int) (int, error)
	GetTotalSKSTx(ctx context.Context, tx pgx.Tx, studentID int, tahunAkademik string) (int, error)
	ExistsTx(ctx context.Context, tx pgx.Tx, studentID, courseID int, tahunAkademik string) (bool, error)
	FindByID(ctx context.Context, id int) (model.Enrollment, error)
	Delete(ctx context.Context, id int) error
}

type enrollmentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentPostgresRepository{pool: pool}
}

func (r *enrollmentPostgresRepository) CreateTx(ctx context.Context, tx pgx.Tx, e model.Enrollment) (model.Enrollment, error) {
	err := tx.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		e.StudentID, e.CourseID, e.TahunAkademik,
	).Scan(&e.ID, &e.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Enrollment{}, ErrDuplicate
		}
		return model.Enrollment{}, fmt.Errorf("menyimpan enrollment: %w", err)
	}
	return e, nil
}

func (r *enrollmentPostgresRepository) CountByCourseTx(ctx context.Context, tx pgx.Tx, courseID int) (int, error) {
	var count int
	err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM enrollments WHERE course_id = $1`, courseID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("menghitung enrollment course: %w", err)
	}
	return count, nil
}

func (r *enrollmentPostgresRepository) GetTotalSKSTx(ctx context.Context, tx pgx.Tx, studentID int, tahunAkademik string) (int, error) {
	var total int
	err := tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahunAkademik,
	).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("menghitung total SKS student: %w", err)
	}
	return total, nil
}

func (r *enrollmentPostgresRepository) ExistsTx(ctx context.Context, tx pgx.Tx, studentID, courseID int, tahunAkademik string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM enrollments
			WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3
		)`, studentID, courseID, tahunAkademik,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("memeriksa keberadaan enrollment: %w", err)
	}
	return exists, nil
}

func (r *enrollmentPostgresRepository) FindByID(ctx context.Context, id int) (model.Enrollment, error) {
	var e model.Enrollment
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, course_id, tahun_akademik, created_at
		 FROM enrollments WHERE id = $1`, id,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("mengambil enrollment: %w", err)
	}
	return e, nil
}

func (r *enrollmentPostgresRepository) Delete(ctx context.Context, id int) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus enrollment: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
