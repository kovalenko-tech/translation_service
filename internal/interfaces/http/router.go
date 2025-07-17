package http

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all API routes
func SetupRoutes(app *fiber.App, handler *Handler, apiKey string) {
	// API v1 group
	api := app.Group("/api/v1")

	// Health check (public endpoint)
	api.Get("/health", handler.HealthCheck)

	// Translation endpoints (protected with API key)
	translations := api.Group("/translations", AuthMiddleware(apiKey))
	translations.Post("/", handler.CreateTranslationRequest)
	translations.Get("/:id", handler.GetTranslationRequest)
	translations.Delete("/:key", handler.DeleteTranslationKey)

	// Swagger documentation with security support
	app.Get("/swagger/*", SwaggerHandler())
}
