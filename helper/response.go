package helper

import (
	"api-students/app/model"
	"github.com/gofiber/fiber/v2"
)

type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    any         `json:"data,omitempty"`
	Meta    *model.Meta `json:"meta,omitempty"`
}

func Ok(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func OkList(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true, Message: message, Data: data, Meta: meta,
	})
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	if location != "" {
		c.Set("Location", location)
	}
	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(WebResponse{
		Success: true, Message: message, Data: data,
	})
}
