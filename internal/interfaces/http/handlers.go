package http

import (
	"fmt"
	"net/http"

	"translation/internal/application/translation"
	domainTranslation "translation/internal/domain/translation"
	"translation/internal/interfaces/http/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler represents HTTP handlers
type Handler struct {
	appService *translation.Service
}

// NewHandler creates a new HTTP handler instance
func NewHandler(appService *translation.Service) *Handler {
	return &Handler{
		appService: appService,
	}
}

// CreateTranslationRequest creates a new translation request
// @Summary Create translation request
// @Description Create a new translation request and queue it for processing
// @Tags translations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.CreateTranslationRequestRequest true "Translation request data"
// @Success 201 {object} dto.CreateTranslationRequestResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/translations [post]
func (h *Handler) CreateTranslationRequest(c *fiber.Ctx) error {
	var req dto.CreateTranslationRequestRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate input data
	if len(req.SourceData) == 0 {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Source data is required",
		})
	}

	if len(req.Languages) == 0 {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "At least one language is required",
		})
	}

	// Create translation request
	request, err := h.appService.CreateTranslationRequest(c.Context(), req.SourceData, req.Languages)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: fmt.Sprintf("Failed to create translation request: %v", err),
		})
	}

	response := dto.CreateTranslationRequestResponse{
		RequestID: request.ID.String(),
		Status:    string(request.Status),
		Message:   "Translation request created successfully and queued for processing",
	}

	return c.Status(http.StatusCreated).JSON(response)
}

// GetTranslationRequest gets request status by ID
// @Summary Get translation request
// @Description Get translation request status and details by ID
// @Tags translations
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Request ID" format(uuid)
// @Success 200 {object} dto.GetTranslationRequestResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/translations/{id} [get]
func (h *Handler) GetTranslationRequest(c *fiber.Ctx) error {
	requestIDStr := c.Params("id")

	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid request ID format",
		})
	}

	request, err := h.appService.GetTranslationRequest(c.Context(), requestID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(dto.ErrorResponse{
			Error: "Translation request not found",
		})
	}

	response := dto.GetTranslationRequestResponse{
		RequestID:  request.ID.String(),
		Status:     string(request.Status),
		SourceData: request.SourceData,
		Languages:  request.Languages,
		CreatedAt:  request.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  request.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if request.CompletedAt != nil {
		completedAt := request.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}

	// Get translated data if request is completed
	if request.Status == domainTranslation.StatusCompleted {
		translatedData, err := h.appService.GetTranslatedDataForRequestKeys(c.Context(), request.SourceData, request.Languages)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to get translated data: %v\n", err)
		} else {
			response.TranslatedData = translatedData
		}
	}

	return c.JSON(response)
}

// HealthCheck checks service status
// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Router /api/v1/health [get]
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(dto.HealthResponse{
		Status:  "ok",
		Message: "Translation service is running",
		Author:  "Kyrylo Kovalenko",
		Contact: "git@kovalenko.tech",
		Website: "https://kovalenko.tech",
	})
}

// DeleteTranslationKey deletes translation key and all its translations
// @Summary Delete translation key
// @Description Delete translation key and all its translations by key
// @Tags translations
// @Security ApiKeyAuth
// @Param key path string true "Translation key"
// @Success 204
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/translations/{key} [delete]
func (h *Handler) DeleteTranslationKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Key is required",
		})
	}

	err := h.appService.DeleteTranslationKey(c.Context(), key)
	if err != nil {
		if err.Error() == "translation key not found" {
			return c.Status(http.StatusNotFound).JSON(dto.ErrorResponse{
				Error: err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.SendStatus(http.StatusNoContent)
}
