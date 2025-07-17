package translation

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines interface for working with translations
type Repository interface {
	// Save translation request
	SaveRequest(ctx context.Context, request *TranslationRequest) error

	// Get request by ID
	GetRequestByID(ctx context.Context, id uuid.UUID) (*TranslationRequest, error)

	// Update request status
	UpdateRequestStatus(ctx context.Context, id uuid.UUID, status RequestStatus) error

	// Save translation key
	SaveTranslationKey(ctx context.Context, key *TranslationKey) error

	// Get translation key
	GetTranslationKey(ctx context.Context, key string) (*TranslationKey, error)

	// Get all translation keys
	GetAllTranslationKeys(ctx context.Context) ([]*TranslationKey, error)

	// Check key existence
	KeyExists(ctx context.Context, key string) (bool, error)

	// Update translation key value and clear translations
	UpdateTranslationKeyValue(ctx context.Context, key string, newValue string) error

	// Delete translation key and all its translations
	DeleteTranslationKey(ctx context.Context, key string) error

	// Get all incomplete requests (pending, processing)
	GetIncompleteRequests(ctx context.Context) ([]*TranslationRequest, error)
}
