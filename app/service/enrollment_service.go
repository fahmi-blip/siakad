package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type EnrollmentService struct {
	pool           *pgxpool.Pool
	studentRepo    repository.StudentRepository
	courseRepo     repository.CourseRepository
	enrollmentRepo repository.EnrollmentRepository
}

func NewEnrollmentService(
	pool *pgxpool.Pool,
	studentRepo repository.StudentRepository,
	courseRepo repository.CourseRepository,
	enrollmentRepo repository.EnrollmentRepository,
) *EnrollmentService {
	return &EnrollmentService{
		pool:           pool,
		studentRepo:    studentRepo,
		courseRepo:     courseRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	if current.Role != "mahasiswa" {
		return helper.Forbidden("Hanya mahasiswa yang dapat mengambil KRS")
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	if err := ValidateTahunAkademik(req.TahunAkademik); err != nil {
		return helper.Validation(map[string][]string{"tahun_akademik": {err.Error()}})
	}

	student, err := s.studentRepo.FindByUserID(ctx, current.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Forbidden("Data mahasiswa Anda tidak ditemukan")
		}
		return helper.Internal(err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	course, err := s.courseRepo.FindByIDTxForUpdate(ctx, tx, req.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Validation(map[string][]string{"course_id": {"Mata kuliah tidak ditemukan"}})
		}
		return helper.Internal(err)
	}

	alreadyEnrolled, err := s.enrollmentRepo.ExistsTx(ctx, tx, student.ID, course.ID, req.TahunAkademik)
	if err != nil {
		return helper.Internal(err)
	}
	if alreadyEnrolled {
		return helper.Conflict("Mata kuliah sudah pernah diambil pada tahun akademik ini")
	}

	terisi, err := s.enrollmentRepo.CountByCourseTx(ctx, tx, course.ID)
	if err != nil {
		return helper.Internal(err)
	}
	if terisi >= course.Kuota {
		return &helper.AppError{
			Status:  fiber.StatusUnprocessableEntity,
			Code:    helper.CodeValidation,
			Message: "kuota penuh",
		}
	}

	totalSKS, err := s.enrollmentRepo.GetTotalSKSTx(ctx, tx, student.ID, req.TahunAkademik)
	if err != nil {
		return helper.Internal(err)
	}

	batasSKS := MaxSKS(student.IpkTerakhir)
	if totalSKS+course.SKS > batasSKS {
		sisaSKS := batasSKS - totalSKS
		if sisaSKS < 0 {
			sisaSKS = 0
		}
		return &helper.AppError{
			Status:  fiber.StatusUnprocessableEntity,
			Code:    helper.CodeValidation,
			Message: fmt.Sprintf("Sisa SKS Anda %d, mata kuliah ini %d SKS", sisaSKS, course.SKS),
		}
	}

	newEnrollment, err := s.enrollmentRepo.CreateTx(ctx, tx, model.Enrollment{
		StudentID:     student.ID,
		CourseID:      course.ID,
		TahunAkademik: req.TahunAkademik,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("Mata kuliah sudah pernah diambil pada tahun akademik ini")
		}
		return helper.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.Internal(err)
	}

	return helper.Created(c, "Mata kuliah berhasil diambil", newEnrollment, "")
}

func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	if current.Role != "mahasiswa" {
		return helper.Forbidden("Hanya mahasiswa yang dapat membatalkan KRS")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.studentRepo.FindByUserID(ctx, current.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Forbidden("Data mahasiswa Anda tidak ditemukan")
		}
		return helper.Internal(err)
	}

	enrollment, err := s.enrollmentRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Enrollment tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if enrollment.StudentID != student.ID {
		return helper.Forbidden("Anda tidak berhak menghapus KRS milik mahasiswa lain")
	}

	if err := s.enrollmentRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Enrollment tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}
