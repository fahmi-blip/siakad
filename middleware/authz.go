package middleware

import (
	"github.com/gofiber/fiber/v2"

	"api-students/helper"
)

func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		if _, granted := allowed[user.Role]; !granted {
			return helper.Forbidden("role anda tidak berhak mengakses endpoint ini")
		}
		return c.Next()
	}
}