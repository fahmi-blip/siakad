package helper

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

const (
	FormatJSON = fiber.MIMEApplicationJSON
	FormatCSV  = "text/csv"
)

func Negotiate(c *fiber.Ctx, offered ...string) (string, error) {
	accept := strings.TrimSpace(c.Get(fiber.HeaderAccept))
	if accept == "" {
		return offered[0], nil
	}
	chosen := c.Accepts(offered...)
	if chosen == "" {
		return "", NotAcceptable(
			"format yang diminta tidak tersedia, pilih salah satu dari: " +
				strings.Join(offered, ", "))
	}
	return chosen, nil
}

func WriteUsersCSV(c *fiber.Ctx, users []model.User) error {
	c.Set(fiber.HeaderContentType, FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="users.csv"`)
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	header := []string{"id", "email", "role", "created_at"}
	if err := writer.Write(header); err != nil {
		return Internal(err)
	}

	for _, u := range users {
		row := []string{
			strconv.Itoa(u.ID), u.Email, u.Role,
			u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
		if err := writer.Write(row); err != nil {
			return Internal(err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return Internal(err)
	}

	return c.SendString(buffer.String())
}

func WriteStudentsCSV(c *fiber.Ctx, students []model.Student) error {
	c.Set(fiber.HeaderContentType, FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="students.csv"`)
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	if err := writer.Write([]string{"id", "user_id", "nim", "nama", "prodi", "angkatan", "ipk_terakhir", "created_at"}); err != nil {
		return Internal(err)
	}
	for _, student := range students {
		if err := writer.Write([]string{
			strconv.Itoa(student.ID),
			strconv.Itoa(student.UserID),
			student.Nim,
			student.Nama,
			student.Prodi,
			strconv.Itoa(student.Angkatan),
			strconv.FormatFloat(student.IpkTerakhir, 'f', 2, 64),
			student.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}); err != nil {
			return Internal(err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return Internal(err)
	}

	return c.SendString(buffer.String())
}
