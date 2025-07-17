package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SwaggerConfig returns swagger configuration with security
func SwaggerConfig() swagger.Config {
	return swagger.Config{
		URL:               "/swagger/doc.json",
		DeepLinking:       true,
		DocExpansion:      "none",
		Title:             "Translation Service API - by Kyrylo Kovalenko",
		OAuth2RedirectUrl: "http://localhost:8080/swagger/oauth2-redirect.html",
		TryItOutEnabled:   true,
	}
}

// SwaggerHandler returns configured swagger handler
func SwaggerHandler() fiber.Handler {
	return swagger.New(SwaggerConfig())
}
