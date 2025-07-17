package translation

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Service represents domain service for working with translations
type Service struct {
	repo Repository
}

// NewService creates a new service instance
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateTranslationRequest creates a new translation request
func (s *Service) CreateTranslationRequest(ctx context.Context, sourceData map[string]string, languages []string) (*TranslationRequest, error) {
	request := NewTranslationRequest(sourceData, languages)

	if err := s.repo.SaveRequest(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to save translation request: %w", err)
	}

	return request, nil
}

// GetTranslationRequest gets request by ID
func (s *Service) GetTranslationRequest(ctx context.Context, id uuid.UUID) (*TranslationRequest, error) {
	return s.repo.GetRequestByID(ctx, id)
}

// ProcessTranslationRequest processes translation request
func (s *Service) ProcessTranslationRequest(ctx context.Context, requestID uuid.UUID) error {
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	// Mark as processing
	request.MarkAsProcessing()
	if err := s.repo.UpdateRequestStatus(ctx, requestID, request.Status); err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	// Extract translation keys from ARB data
	translationKeys, err := s.extractTranslationKeys(request.SourceData)
	if err != nil {
		request.MarkAsFailed()
		s.repo.UpdateRequestStatus(ctx, requestID, request.Status)
		return fmt.Errorf("failed to extract translation keys: %w", err)
	}

	// Process each key - check if it exists and if value has changed
	for _, newKey := range translationKeys {
		existingKey, err := s.repo.GetTranslationKey(ctx, newKey.Key)
		if err != nil {
			// Key doesn't exist, save as new
			if err := s.repo.SaveTranslationKey(ctx, newKey); err != nil {
				// Log error but continue processing
				fmt.Printf("Failed to save new translation key %s: %v\n", newKey.Key, err)
			}
			continue
		}

		// Key exists, check if value has changed
		if existingKey.Value != newKey.Value {
			// Value has changed, update the key and clear existing translations
			// so they will be regenerated
			if err := s.repo.UpdateTranslationKeyValue(ctx, newKey.Key, newKey.Value); err != nil {
				// Log error but continue processing
				fmt.Printf("Failed to update translation key %s with new value: %v\n", newKey.Key, err)
			} else {
				fmt.Printf("Updated translation key %s with new value: %s -> %s\n", newKey.Key, existingKey.Value, newKey.Value)
			}
		}
		// If value hasn't changed, keep existing translations
	}

	// Don't mark as completed here - let the application service do it after translations
	return nil
}

// extractTranslationKeys extracts translation keys from ARB data
func (s *Service) extractTranslationKeys(sourceData map[string]string) ([]*TranslationKey, error) {
	var keys []*TranslationKey

	for key, value := range sourceData {
		// Skip keys starting with @ or @@
		if strings.HasPrefix(key, "@") {
			continue
		}

		// Create translation key for string value
		translationKey := &TranslationKey{
			Key:          key,
			Value:        value,
			Translations: make(map[string]string),
		}
		keys = append(keys, translationKey)
	}

	return keys, nil
}

// GetPendingTranslationKeys gets keys that require translation
func (s *Service) GetPendingTranslationKeys(ctx context.Context, languages []string) ([]*TranslationKey, error) {
	allKeys, err := s.repo.GetAllTranslationKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all translation keys: %w", err)
	}

	var pendingKeys []*TranslationKey
	for _, key := range allKeys {
		needsTranslation := false
		for _, lang := range languages {
			if _, exists := key.Translations[lang]; !exists {
				needsTranslation = true
				break
			}
		}

		if needsTranslation {
			pendingKeys = append(pendingKeys, key)
		}
	}

	return pendingKeys, nil
}

// GetPendingTranslationKeysForRequest gets keys that require translation for specific request keys and languages
func (s *Service) GetPendingTranslationKeysForRequest(ctx context.Context, sourceData map[string]string, languages []string) ([]*TranslationKey, error) {
	var pendingKeys []*TranslationKey

	// Process each key from the request
	for keyName, keyValue := range sourceData {
		// Skip keys starting with @
		if strings.HasPrefix(keyName, "@") {
			continue
		}

		// Get existing key or create new one
		existingKey, err := s.repo.GetTranslationKey(ctx, keyName)
		if err != nil {
			// Key doesn't exist, create new one
			newKey := &TranslationKey{
				Key:          keyName,
				Value:        keyValue,
				Translations: make(map[string]string),
			}
			pendingKeys = append(pendingKeys, newKey)
			continue
		}

		// Check if this key needs translation for any of the requested languages
		needsTranslation := false
		for _, lang := range languages {
			if _, exists := existingKey.Translations[lang]; !exists {
				needsTranslation = true
				break
			}
		}

		// Only add to pending if translations are missing
		if needsTranslation {
			// Update the value if it's different (but don't clear existing translations)
			if existingKey.Value != keyValue {
				existingKey.Value = keyValue
			}
			pendingKeys = append(pendingKeys, existingKey)
		} else {
			// All translations exist, just update the value if needed
			if existingKey.Value != keyValue {
				existingKey.Value = keyValue
				// Save the updated value without triggering translation
				if err := s.repo.SaveTranslationKey(ctx, existingKey); err != nil {
					fmt.Printf("Failed to update value for key %s: %v\n", keyName, err)
				}
			}
			fmt.Printf("Key %s already has all required translations, skipping translation process\n", keyName)
		}
	}

	return pendingKeys, nil
}

// GetRepository returns repository
func (s *Service) GetRepository() Repository {
	return s.repo
}

// GetTranslatedData gets translated data for specific languages
func (s *Service) GetTranslatedData(ctx context.Context, languages []string) (map[string]map[string]string, error) {
	allKeys, err := s.repo.GetAllTranslationKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all translation keys: %w", err)
	}

	translatedData := make(map[string]map[string]string)

	// Initialize language maps
	for _, lang := range languages {
		translatedData[lang] = make(map[string]string)
	}

	// Fill translated data
	for _, key := range allKeys {
		for _, lang := range languages {
			if translation, exists := key.Translations[lang]; exists {
				translatedData[lang][key.Key] = translation
			}
		}
	}

	return translatedData, nil
}

// GetTranslatedDataForRequestKeys gets translated data only for keys that were in the request's source_data
func (s *Service) GetTranslatedDataForRequestKeys(ctx context.Context, requestKeys map[string]string, languages []string) (map[string]map[string]string, error) {
	translatedData := make(map[string]map[string]string)

	// Initialize language maps
	for _, lang := range languages {
		translatedData[lang] = make(map[string]string)
	}

	// Get translations only for keys that were in the request
	for keyName := range requestKeys {
		key, err := s.repo.GetTranslationKey(ctx, keyName)
		if err != nil {
			// Skip keys that can't be found
			continue
		}

		for _, lang := range languages {
			if translation, exists := key.Translations[lang]; exists {
				translatedData[lang][key.Key] = translation
			}
		}
	}

	return translatedData, nil
}

// CancelTranslationRequest cancels a translation request
func (s *Service) CancelTranslationRequest(ctx context.Context, requestID uuid.UUID) error {
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	// Check if request can be cancelled
	if request.Status == StatusCompleted || request.Status == StatusFailed || request.Status == StatusCancelled {
		return fmt.Errorf("request cannot be cancelled in status: %s", request.Status)
	}

	// Mark as cancelled
	request.MarkAsCancelled()
	return s.repo.UpdateRequestStatus(ctx, requestID, request.Status)
}

// CompleteTranslationRequest marks a translation request as completed
func (s *Service) CompleteTranslationRequest(ctx context.Context, requestID uuid.UUID) error {
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	// Mark as completed
	request.MarkAsCompleted()
	return s.repo.UpdateRequestStatus(ctx, requestID, request.Status)
}

// GetIncompleteRequests gets all requests that are not completed, failed, or cancelled
func (s *Service) GetIncompleteRequests(ctx context.Context) ([]*TranslationRequest, error) {
	return s.repo.GetIncompleteRequests(ctx)
}

// DeleteTranslationKey deletes translation key and all its translations
func (s *Service) DeleteTranslationKey(ctx context.Context, key string) error {
	// Check if key exists
	exists, err := s.repo.KeyExists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check key existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("translation key not found")
	}

	// Delete the key and all its translations
	if err := s.repo.DeleteTranslationKey(ctx, key); err != nil {
		return fmt.Errorf("failed to delete translation key: %w", err)
	}

	return nil
}

// CacheTranslationsResult represents the result of caching translations
type CacheTranslationsResult struct {
	SuccessCount int
	SkippedKeys  []string
	TotalKeys    int
}

// CacheTranslations caches translations for keys without running translation process
func (s *Service) CacheTranslations(ctx context.Context, translations map[string]map[string]string) (*CacheTranslationsResult, error) {
	result := &CacheTranslationsResult{
		SkippedKeys: []string{},
	}

	// First, collect all keys and their translations
	keyTranslations := make(map[string]map[string]string)

	// Process each language
	for lang, langTranslations := range translations {
		for keyName, translationValue := range langTranslations {
			if keyTranslations[keyName] == nil {
				keyTranslations[keyName] = make(map[string]string)
			}
			keyTranslations[keyName][lang] = translationValue
		}
	}

	result.TotalKeys = len(keyTranslations)

	// Now process each key
	for keyName, langTranslations := range keyTranslations {
		// Check if we have English translation (source language)
		englishValue, hasEnglish := langTranslations["en"]
		if !hasEnglish {
			// Skip keys without English translation
			fmt.Printf("Skipping key %s - no English translation provided\n", keyName)
			result.SkippedKeys = append(result.SkippedKeys, keyName)
			continue
		}

		// Get existing key or create new one
		existingKey, err := s.repo.GetTranslationKey(ctx, keyName)
		if err != nil {
			// Key doesn't exist, create new one with English value
			newKey := &TranslationKey{
				Key:          keyName,
				Value:        englishValue, // Use English translation as source value
				Translations: make(map[string]string),
			}

			// Add all translations
			for lang, translationValue := range langTranslations {
				newKey.Translations[lang] = translationValue
			}

			if err := s.repo.SaveTranslationKey(ctx, newKey); err != nil {
				return result, fmt.Errorf("failed to save new translation key %s: %w", keyName, err)
			}
			result.SuccessCount++
		} else {
			// Key exists, update translations and value
			// Update the value to English translation
			existingKey.Value = englishValue

			// Add or update all translations
			for lang, translationValue := range langTranslations {
				existingKey.Translations[lang] = translationValue
			}

			if err := s.repo.SaveTranslationKey(ctx, existingKey); err != nil {
				return result, fmt.Errorf("failed to update translation key %s: %w", keyName, err)
			}
			result.SuccessCount++
		}
	}

	return result, nil
}
