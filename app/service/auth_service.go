package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type AuthService struct {
	users    repository.UserRepository
	students repository.StudentRepository
	jwt      *helper.JWTManager
}

func NewAuthService(
	users repository.UserRepository,
	students repository.StudentRepository,
	jwtManager *helper.JWTManager,
) *AuthService {
	return &AuthService{
		users:    users,
		students: students,
		jwt:      jwtManager,
	}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Email = strings.TrimSpace(req.Email)

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		helper.VerifyDummyPassword(req.Password)
		return helper.Unauthorized("kredensial salah")
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Unauthorized("kredensial salah")
	}

	// Mahasiswa yang sudah soft delete TIDAK boleh bisa login
	if user.Role == "mahasiswa" {
		student, err := s.students.FindByUserID(ctx, user.ID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return helper.Unauthorized("kredensial salah")
			}
			return helper.Internal(err)
		}
		_ = student
	}

	token, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return helper.Internal(err)
	}

	resp := model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwt.AccessTTL().Seconds()),
		User: model.AuthUserInfo{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
	}

	return helper.Ok(c, fiber.StatusOK, "login berhasil", resp)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		return helper.Unauthorized("user tidak ditemukan")
	}

	meResp := model.MeResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	if user.Role == "mahasiswa" {
		student, err := s.students.FindByUserID(ctx, user.ID)
		if err == nil {
			meResp.Student = &model.MeStudentData{
				ID:          student.ID,
				Nim:         student.Nim,
				Nama:        student.Nama,
				Prodi:       student.Prodi,
				Angkatan:    student.Angkatan,
				IpkTerakhir: student.IpkTerakhir,
			}
		}
	}

	return helper.Ok(c, fiber.StatusOK, "profil berhasil diambil", meResp)
}
