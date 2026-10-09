package service

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type CourseService struct {
	repo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	q := model.CourseQuery{
		Search:    strings.TrimSpace(c.Query("search")),
		Available: c.QueryBool("available", false),
	}

	if semesterStr := c.Query("semester"); semesterStr != "" {
		semester := c.QueryInt("semester", 0)
		if semester > 0 {
			q.Semester = &semester
		}
	}

	courses, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Ok(c, fiber.StatusOK, "Daftar mata kuliah berhasil diambil", courses)
}
