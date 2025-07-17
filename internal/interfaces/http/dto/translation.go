package dto

// CreateTranslationRequestRequest represents translation creation request
type CreateTranslationRequestRequest struct {
	SourceData map[string]string `json:"source_data" validate:"required" example:{"hello":"Hello World","welcome":"Welcome to our app","goodbye":"Goodbye"}`
	Languages  []string          `json:"languages" example:"es,fr,de" validate:"required,min=1"`
}

// CreateTranslationRequestResponse represents response to creation request
type CreateTranslationRequestResponse struct {
	RequestID string `json:"request_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status    string `json:"status" example:"pending"`
	Message   string `json:"message" example:"Translation request created successfully and queued for processing"`
}

// GetTranslationRequestResponse represents response to get request
type GetTranslationRequestResponse struct {
	RequestID      string                       `json:"request_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status         string                       `json:"status" example:"completed"`
	SourceData     map[string]string            `json:"source_data" example:{"hello":"Hello World"}`
	Languages      []string                     `json:"languages" example:"es,fr,de"`
	TranslatedData map[string]map[string]string `json:"translated_data,omitempty" example:{"es":{"hello":"Hola Mundo","welcome":"Bienvenido a nuestra aplicación"},"fr":{"hello":"Bonjour le monde","welcome":"Bienvenue dans notre application"}}`
	CreatedAt      string                       `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt      string                       `json:"updated_at" example:"2024-01-01T12:05:00Z"`
	CompletedAt    *string                      `json:"completed_at,omitempty" example:"2024-01-01T12:05:00Z"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request body"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message" example:"Translation service is running"`
	Author  string `json:"author" example:"Kyrylo Kovalenko"`
	Contact string `json:"contact" example:"git@kovalenko.tech"`
	Website string `json:"website" example:"https://kovalenko.tech"`
}
