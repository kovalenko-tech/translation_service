package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"translation/internal/domain/translation"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Repository implements repository interface for Redis
type Repository struct {
	client *redis.Client
}

// NewRepository creates a new Redis repository instance
func NewRepository(client *redis.Client) *Repository {
	return &Repository{
		client: client,
	}
}

// SaveRequest saves translation request to Redis
func (r *Repository) SaveRequest(ctx context.Context, request *translation.TranslationRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	key := fmt.Sprintf("translation_request:%s", request.ID.String())
	return r.client.Set(ctx, key, data, 24*time.Hour).Err()
}

// GetRequestByID gets request by ID from Redis
func (r *Repository) GetRequestByID(ctx context.Context, id uuid.UUID) (*translation.TranslationRequest, error) {
	key := fmt.Sprintf("translation_request:%s", id.String())
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("request not found")
		}
		return nil, fmt.Errorf("failed to get request: %w", err)
	}

	var request translation.TranslationRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return &request, nil
}

// UpdateRequestStatus updates request status in Redis
func (r *Repository) UpdateRequestStatus(ctx context.Context, id uuid.UUID, status translation.RequestStatus) error {
	request, err := r.GetRequestByID(ctx, id)
	if err != nil {
		return err
	}

	request.Status = status
	request.UpdatedAt = time.Now()

	return r.SaveRequest(ctx, request)
}

// SaveTranslationKey saves translation key to Redis
func (r *Repository) SaveTranslationKey(ctx context.Context, key *translation.TranslationKey) error {
	data, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal translation key: %w", err)
	}

	redisKey := fmt.Sprintf("translation_key:%s", key.Key)
	return r.client.Set(ctx, redisKey, data, 0).Err() // No TTL for permanent storage
}

// GetTranslationKey gets translation key from Redis
func (r *Repository) GetTranslationKey(ctx context.Context, key string) (*translation.TranslationKey, error) {
	redisKey := fmt.Sprintf("translation_key:%s", key)
	data, err := r.client.Get(ctx, redisKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("translation key not found")
		}
		return nil, fmt.Errorf("failed to get translation key: %w", err)
	}

	var translationKey translation.TranslationKey
	if err := json.Unmarshal(data, &translationKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal translation key: %w", err)
	}

	return &translationKey, nil
}

// GetAllTranslationKeys gets all translation keys from Redis
func (r *Repository) GetAllTranslationKeys(ctx context.Context) ([]*translation.TranslationKey, error) {
	pattern := "translation_key:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get translation keys: %w", err)
	}

	var translationKeys []*translation.TranslationKey
	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue // Skip problematic keys
		}

		var translationKey translation.TranslationKey
		if err := json.Unmarshal(data, &translationKey); err != nil {
			continue // Skip problematic keys
		}

		translationKeys = append(translationKeys, &translationKey)
	}

	return translationKeys, nil
}

// KeyExists checks key existence in Redis
func (r *Repository) KeyExists(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("translation_key:%s", key)
	exists, err := r.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return exists > 0, nil
}

// UpdateTranslationKeyValue updates translation key value and clears existing translations
func (r *Repository) UpdateTranslationKeyValue(ctx context.Context, key string, newValue string) error {
	// Get existing key
	existingKey, err := r.GetTranslationKey(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get existing translation key: %w", err)
	}

	// Update value and clear translations
	existingKey.Value = newValue
	existingKey.Translations = make(map[string]string)

	// Save updated key
	return r.SaveTranslationKey(ctx, existingKey)
}

// DeleteTranslationKey deletes translation key and all its translations from Redis
func (r *Repository) DeleteTranslationKey(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("translation_key:%s", key)

	// Check if key exists
	exists, err := r.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check key existence: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("translation key not found")
	}

	// Delete the key
	return r.client.Del(ctx, redisKey).Err()
}
