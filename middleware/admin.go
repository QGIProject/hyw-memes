package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func AdminAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		adminToken := c.Get("X-Admin-Token")
		if adminToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Admin authentication required",
			})
		}

		// Validate admin session (stored in locals during login)
		session := c.Cookies("admin_session")
		if session == "" || session != "authenticated" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid admin session",
			})
		}

		return c.Next()
	}
}
