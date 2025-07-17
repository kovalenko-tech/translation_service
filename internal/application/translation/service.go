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

	// Process request in domain
	if err := s.domainService.ProcessTranslationRequest(ctx, task.RequestID); err != nil {
		return fmt.Errorf("failed to process translation request: %w", err)
	}

	// Get keys that require translation for the specific request keys and languages
	pendingKeys, err := s.domainService.GetPendingTranslationKeysForRequest(ctx, task.SourceData, task.Languages)
	if err != nil {
		return fmt.Errorf("failed to get pending translation keys: %w", err)
	}

	if len(pendingKeys) == 0 {
		log.Printf("No pending translation keys found for request ID: %s", task.RequestID)
		return nil
	}

	// Generate translations for each key
	for _, key := range pendingKeys {
		if err := s.translateKey(ctx, key, task.Languages); err != nil {
			log.Printf("Failed to translate key %s: %v", key.Key, err)
			continue
		}

		// Save updated key
		if err := s.domainService.GetRepository().SaveTranslationKey(ctx, key); err != nil {
			log.Printf("Failed to save translated key %s: %v", key.Key, err)
		}
	}

	log.Printf("Successfully processed translation task for request ID: %s", task.RequestID)
	return nil
}

// translateKey translates one key to all specified languages
func (s *Service) translateKey(ctx context.Context, key *translation.TranslationKey, languages []string) error {
	// Assume source language is English (can be made configurable)
	sourceLanguage := "en"

	for _, targetLang := range languages {
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
