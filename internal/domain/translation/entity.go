package translation

import (
	"time"

	"github.com/google/uuid"
)

// TranslationRequest represents translation request
type TranslationRequest struct {
	ID          uuid.UUID         `json:"id"`
	Status      RequestStatus     `json:"status"`
	SourceData  map[string]string `json:"source_data"`
	Languages   []string          `json:"languages"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

// TranslationKey represents translation key
type TranslationKey struct {
	Key          string            `json:"key"`
	Value        string            `json:"value"`
	Translations map[string]string `json:"translations"`
}

// RequestStatus represents request status
type RequestStatus string

const (
	StatusPending    RequestStatus = "pending"
	StatusProcessing RequestStatus = "processing"
	StatusCompleted  RequestStatus = "completed"
	StatusFailed     RequestStatus = "failed"
)

// NewTranslationRequest creates a new translation request
func NewTranslationRequest(sourceData map[string]string, languages []string) *TranslationRequest {
	return &TranslationRequest{
		ID:         uuid.New(),
		Status:     StatusPending,
		SourceData: sourceData,
		Languages:  languages,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// MarkAsProcessing marks request as processing
func (tr *TranslationRequest) MarkAsProcessing() {
	tr.Status = StatusProcessing
	tr.UpdatedAt = time.Now()
}

// MarkAsCompleted marks request as completed
func (tr *TranslationRequest) MarkAsCompleted() {
	tr.Status = StatusCompleted
	tr.UpdatedAt = time.Now()
	now := time.Now()
	tr.CompletedAt = &now
}

// MarkAsFailed marks request as failed
func (tr *TranslationRequest) MarkAsFailed() {
	tr.Status = StatusFailed
	tr.UpdatedAt = time.Now()
}
