package translation

import (
	"context"
	"fmt"
	"log"
	"strings"

	"translation/internal/domain/translation"
	"translation/internal/infrastructure/openai"
	"translation/internal/infrastructure/rabbitmq"

	"github.com/google/uuid"
)

// Service represents application service for working with translations
type Service struct {
	domainService *translation.Service
	openaiService *openai.Service
	rabbitService *rabbitmq.Service
}

// NewService creates a new application service instance
func NewService(
	domainService *translation.Service,
	openaiService *openai.Service,
	rabbitService *rabbitmq.Service,
) *Service {
	return &Service{
		domainService: domainService,
		openaiService: openaiService,
		rabbitService: rabbitService,
	}
}

// CreateTranslationRequest creates a new translation request and sends it to the queue
func (s *Service) CreateTranslationRequest(ctx context.Context, sourceData map[string]string, languages []string) (*translation.TranslationRequest, error) {
	// Create request in domain
	request, err := s.domainService.CreateTranslationRequest(ctx, sourceData, languages)
	if err != nil {
		return nil, fmt.Errorf("failed to create translation request: %w", err)
	}

	// Create task for RabbitMQ
	task := &rabbitmq.TranslationTask{
		RequestID:  request.ID,
		SourceData: sourceData,
		Languages:  languages,
	}

	// Send task to queue
	if err := s.rabbitService.PublishTask(ctx, task); err != nil {
		// If failed to send to queue, mark request as failed
		request.MarkAsFailed()
		s.domainService.GetRepository().UpdateRequestStatus(ctx, request.ID, request.Status)
		return nil, fmt.Errorf("failed to publish task to queue: %w", err)
	}

	return request, nil
}

// GetTranslationRequest gets request by ID
func (s *Service) GetTranslationRequest(ctx context.Context, id uuid.UUID) (*translation.TranslationRequest, error) {
	return s.domainService.GetTranslationRequest(ctx, id)
}

// GetTranslatedData gets translated data for specific languages
func (s *Service) GetTranslatedData(ctx context.Context, languages []string) (map[string]map[string]string, error) {
	return s.domainService.GetTranslatedData(ctx, languages)
}

// GetTranslatedDataForRequestKeys gets translated data only for keys that were in the request's source_data
func (s *Service) GetTranslatedDataForRequestKeys(ctx context.Context, requestKeys map[string]string, languages []string) (map[string]map[string]string, error) {
	return s.domainService.GetTranslatedDataForRequestKeys(ctx, requestKeys, languages)
}

// ProcessTranslationTask processes translation task (called by consumer)
func (s *Service) ProcessTranslationTask(ctx context.Context, task *rabbitmq.TranslationTask) error {
	log.Printf("Starting to process translation task for request ID: %s", task.RequestID)

	// Check if request was cancelled before starting
	request, err := s.domainService.GetTranslationRequest(ctx, task.RequestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	if request.Status == translation.StatusCancelled {
		log.Printf("Request %s was cancelled, skipping processing", task.RequestID)
		return nil
	}

	// Process request in domain (this will mark as processing if not already)
	if err := s.domainService.ProcessTranslationRequest(ctx, task.RequestID); err != nil {
		return fmt.Errorf("failed to process translation request: %w", err)
	}

	// Get keys that require translation for the specific request keys and languages
	pendingKeys, err := s.domainService.GetPendingTranslationKeysForRequest(ctx, task.SourceData, task.Languages)
	if err != nil {
		return fmt.Errorf("failed to get pending translation keys: %w", err)
	}

	if len(pendingKeys) == 0 {
		log.Printf("No pending translation keys found for request ID: %s - all translations already exist in cache", task.RequestID)
		// Mark as completed since no translations needed
		if err := s.CompleteTranslationRequest(ctx, task.RequestID); err != nil {
			log.Printf("Failed to mark request as completed: %v", err)
		}
		return nil
	}

	log.Printf("Found %d keys that need translation for request ID: %s", len(pendingKeys), task.RequestID)

	// Generate translations for each key
	for i, key := range pendingKeys {
		// Check if request was cancelled before processing each key
		request, err := s.domainService.GetTranslationRequest(ctx, task.RequestID)
		if err != nil {
			log.Printf("Failed to get request status for ID %s: %v", task.RequestID, err)
			continue
		}

		if request.Status == translation.StatusCancelled {
			log.Printf("Request %s was cancelled, stopping translation process", task.RequestID)
			return nil
		}

		log.Printf("Translating key %d/%d: %s for request ID: %s", i+1, len(pendingKeys), key.Key, task.RequestID)

		if err := s.translateKey(ctx, key, task.Languages, task.RequestID); err != nil {
			if err.Error() == "request was cancelled" {
				log.Printf("Translation cancelled for request %s", task.RequestID)
				return nil
			}
			log.Printf("Failed to translate key %s: %v", key.Key, err)
			continue
		}

		// Save updated key
		if err := s.domainService.GetRepository().SaveTranslationKey(ctx, key); err != nil {
			log.Printf("Failed to save translated key %s: %v", key.Key, err)
		}
	}

	// Mark as completed after all translations are done
	if err := s.CompleteTranslationRequest(ctx, task.RequestID); err != nil {
		log.Printf("Failed to mark request as completed: %v", err)
	}

	log.Printf("Successfully processed translation task for request ID: %s", task.RequestID)
	return nil
}

// translateKey translates one key to all specified languages
func (s *Service) translateKey(ctx context.Context, key *translation.TranslationKey, languages []string, requestID uuid.UUID) error {
	// Assume source language is English (can be made configurable)
	sourceLanguage := "en"

	for _, targetLang := range languages {
		// Check if request was cancelled before each translation
		request, err := s.domainService.GetTranslationRequest(ctx, requestID)
		if err != nil {
			log.Printf("Failed to get request status for ID %s: %v", requestID, err)
			continue
		}

		if request.Status == translation.StatusCancelled {
			log.Printf("Request %s was cancelled, stopping translation of key %s", requestID, key.Key)
			return fmt.Errorf("request was cancelled")
		}

		// Skip if translation already exists
		if _, exists := key.Translations[targetLang]; exists {
			log.Printf("Translation for key %s to %s already exists, skipping", key.Key, targetLang)
			continue
		}

		// Create translation request
		translationReq := &openai.TranslationRequest{
			Text:     key.Value,
			FromLang: sourceLanguage,
			ToLang:   targetLang,
			Context:  fmt.Sprintf("Translation key: %s", key.Key),
		}

		// Perform translation
		resp, err := s.openaiService.Translate(ctx, translationReq)
		if err != nil {
			log.Printf("Failed to translate key %s to %s: %v", key.Key, targetLang, err)
			continue
		}

		// Clean up the translated text - remove extra quotes
		translatedText := resp.TranslatedText
		translatedText = strings.Trim(translatedText, `"'`)

		// Save translation
		key.Translations[targetLang] = translatedText
		log.Printf("Translated key %s to %s: %s -> %s", key.Key, targetLang, key.Value, translatedText)
	}

	return nil
}

// StartConsumer starts consumer for task processing
func (s *Service) StartConsumer(ctx context.Context) error {
	return s.rabbitService.ConsumeTasks(ctx, func(task *rabbitmq.TranslationTask) error {
		return s.ProcessTranslationTask(ctx, task)
	})
}

// DeleteTranslationKey deletes translation key and all its translations
func (s *Service) DeleteTranslationKey(ctx context.Context, key string) error {
	return s.domainService.DeleteTranslationKey(ctx, key)
}

// CacheTranslations caches translations for keys without running translation process
func (s *Service) CacheTranslations(ctx context.Context, translations map[string]map[string]string) (*translation.CacheTranslationsResult, error) {
	return s.domainService.CacheTranslations(ctx, translations)
}

// CancelTranslationRequest cancels a translation request
func (s *Service) CancelTranslationRequest(ctx context.Context, requestID uuid.UUID) error {
	return s.domainService.CancelTranslationRequest(ctx, requestID)
}

// CompleteTranslationRequest marks a translation request as completed
func (s *Service) CompleteTranslationRequest(ctx context.Context, requestID uuid.UUID) error {
	return s.domainService.CompleteTranslationRequest(ctx, requestID)
}

// GetIncompleteRequests gets all requests that are not completed, failed, or cancelled
func (s *Service) GetIncompleteRequests(ctx context.Context) ([]*translation.TranslationRequest, error) {
	return s.domainService.GetIncompleteRequests(ctx)
}

// RecoverIncompleteRequests recovers and processes incomplete requests on server startup
func (s *Service) RecoverIncompleteRequests(ctx context.Context) error {
	log.Printf("Starting recovery of incomplete translation requests...")

	incompleteRequests, err := s.GetIncompleteRequests(ctx)
	if err != nil {
		return fmt.Errorf("failed to get incomplete requests: %w", err)
	}

	if len(incompleteRequests) == 0 {
		log.Printf("No incomplete requests found")
		return nil
	}

	log.Printf("Found %d incomplete requests to recover", len(incompleteRequests))

	for _, request := range incompleteRequests {
		log.Printf("Recovering request ID: %s (status: %s)", request.ID, request.Status)

		// Create task for RabbitMQ
		task := &rabbitmq.TranslationTask{
			RequestID:  request.ID,
			SourceData: request.SourceData,
			Languages:  request.Languages,
		}

		// Send task to queue
		if err := s.rabbitService.PublishTask(ctx, task); err != nil {
			log.Printf("Failed to publish recovery task for request ID %s: %v", request.ID, err)
			// Mark as failed if we can't queue it
			request.MarkAsFailed()
			s.domainService.GetRepository().UpdateRequestStatus(ctx, request.ID, request.Status)
			continue
		}

		log.Printf("Successfully queued recovery task for request ID: %s", request.ID)
	}

	log.Printf("Recovery of incomplete requests completed")
	return nil
}
