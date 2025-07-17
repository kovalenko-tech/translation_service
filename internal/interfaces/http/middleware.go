package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware creates middleware for API key authentication
func AuthMiddleware(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")

		// Check if header exists
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		var token string

		// Check if header starts with "Bearer "
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// If no Bearer prefix, use the entire header value as token
			token = authHeader
		}

		// Validate token
		if token != apiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		// Continue to next handler
		return c.Next()
	}
}
