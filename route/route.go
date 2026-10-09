package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	JWT               *helper.JWTManager
	AuthService       *service.AuthService
	StudentService    *service.StudentService
	CourseService     *service.CourseService
	EnrollmentService *service.EnrollmentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool))

	// 1. Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", middleware.RequireJSON, middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// 2. Students routes
	students := api.Group("/students", middleware.RequireAuth(deps.JWT))
	students.Get("/", middleware.RequireRole("admin"), deps.StudentService.List)
	students.Post("/", middleware.RequireJSON, middleware.RequireRole("admin"), deps.StudentService.Create)
	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", middleware.RequireJSON, middleware.RequireRole("admin"), deps.StudentService.Update)
	students.Delete("/:id", middleware.RequireRole("admin"), deps.StudentService.Delete)

	// Legacy prestasi route
	students.Get("/:id/prestasi", deps.StudentService.GetWithPrestation)

	// 3. Courses routes
	courses := api.Group("/courses", middleware.RequireAuth(deps.JWT))
	courses.Get("/", deps.CourseService.List)

	// 4. Enrollments routes
	enrollments := api.Group("/enrollments", middleware.RequireAuth(deps.JWT))
	enrollments.Post("/", middleware.RequireJSON, middleware.RequireRole("mahasiswa"), deps.EnrollmentService.Create)
	enrollments.Delete("/:id", middleware.RequireRole("mahasiswa"), deps.EnrollmentService.Delete)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
