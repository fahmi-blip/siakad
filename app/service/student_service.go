package service

import (
	"errors"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type StudentService struct {
	pool        *pgxpool.Pool
	userRepo    repository.UserRepository
	studentRepo repository.StudentRepository
}

func NewStudentService(
	pool *pgxpool.Pool,
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
) *StudentService {
	return &StudentService{
		pool:        pool,
		userRepo:    userRepo,
		studentRepo: studentRepo,
	}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	perPage := c.QueryInt("per_page", 10)
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}

	q := model.StudentListQuery{
		Page:     page,
		PerPage:  perPage,
		Prodi:    c.Query("prodi"),
		Angkatan: c.QueryInt("angkatan", 0),
		Search:   c.Query("search"),
		Sort:     c.Query("sort"),
	}

	students, total, err := s.studentRepo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	meta := &model.Meta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	}

	return helper.OkList(c, "Data mahasiswa berhasil diambil", students, meta)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	if err := ValidateAngkatan(req.Angkatan); err != nil {
		return helper.Validation(map[string][]string{"angkatan": {err.Error()}})
	}

	ipk := 0.00
	if req.IpkTerakhir != nil {
		ipk = *req.IpkTerakhir
	}

	hashedPassword, err := helper.HashPassword(req.Nim)
	if err != nil {
		return helper.Internal(err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	newUser, err := s.userRepo.CreateTx(ctx, tx, model.User{
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "mahasiswa",
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Validation(map[string][]string{"email": {"email sudah terdaftar"}})
		}
		return helper.Internal(err)
	}

	newStudent, err := s.studentRepo.CreateTx(ctx, tx, model.Student{
		UserID:      newUser.ID,
		Nim:         req.Nim,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IpkTerakhir: ipk,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Validation(map[string][]string{"nim": {"NIM sudah terdaftar"}})
		}
		return helper.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return helper.Internal(err)
	}

	return helper.Created(c, "Mahasiswa berhasil ditambahkan", newStudent, "")
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	detail, err := s.studentRepo.FindByIDWithCourses(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	detail.BatasSKS = MaxSKS(detail.IpkTerakhir)

	if current.Role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(ctx, current.UserID)
		if err != nil || student.ID != detail.ID {
			return helper.Forbidden("Anda tidak berhak mengakses data mahasiswa lain")
		}
	}

	return helper.Ok(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", detail)
}

func (s *StudentService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	if err := ValidateAngkatan(req.Angkatan); err != nil {
		return helper.Validation(map[string][]string{"angkatan": {err.Error()}})
	}

	existing, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	ipk := existing.IpkTerakhir
	if req.IpkTerakhir != nil {
		ipk = *req.IpkTerakhir
	}

	updated, err := s.studentRepo.Update(ctx, model.Student{
		ID:          id,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IpkTerakhir: ipk,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Ok(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", updated)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.studentRepo.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}

func (s *StudentService) GetWithPrestation(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := s.studentRepo.FindByIDWithPrestasi(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if current.Role == "mahasiswa" {
		ownStudent, err := s.studentRepo.FindByUserID(ctx, current.UserID)
		if err != nil || ownStudent.ID != student.ID {
			return helper.Forbidden("Anda tidak berhak mengakses data mahasiswa ini")
		}
	}

	return helper.Ok(c, fiber.StatusOK, "Data mahasiswa dengan prestasi ditemukan", student)
}
