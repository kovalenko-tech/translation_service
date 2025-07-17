package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// Service represents service for working with OpenAI
type Service struct {
	client *openai.Client
}

// NewService creates a new OpenAI service instance
func NewService(apiKey string) *Service {
	return &Service{
		client: openai.NewClient(apiKey),
	}
}

// TranslationRequest represents translation request
type TranslationRequest struct {
	Text     string `json:"text"`
	FromLang string `json:"from_lang"`
	ToLang   string `json:"to_lang"`
	Context  string `json:"context,omitempty"`
}

// TranslationResponse represents translation response
type TranslationResponse struct {
	TranslatedText string  `json:"translated_text"`
	Confidence     float64 `json:"confidence"`
}

// Translate translates text using OpenAI
func (s *Service) Translate(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error) {
	prompt := s.buildTranslationPrompt(req)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a professional translator. Translate the given text accurately while preserving the meaning and context. Return only the translated text without any additional explanations, formatting, or quotes.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.3, // Low temperature for more consistent translations
			MaxTokens:   1000,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	translatedText := strings.TrimSpace(resp.Choices[0].Message.Content)

	return &TranslationResponse{
		TranslatedText: translatedText,
		Confidence:     0.9, // OpenAI doesn't provide confidence score, use fixed value
	}, nil
}

// TranslateBatch translates multiple texts simultaneously
func (s *Service) TranslateBatch(ctx context.Context, requests []*TranslationRequest) ([]*TranslationResponse, error) {
	var responses []*TranslationResponse

	for _, req := range requests {
		resp, err := s.Translate(ctx, req)
		if err != nil {
			// Return error for specific request but continue processing others
			fmt.Printf("Failed to translate text '%s': %v\n", req.Text, err)
			responses = append(responses, &TranslationResponse{
				TranslatedText: req.Text, // Return original text in case of error
				Confidence:     0.0,
			})
			continue
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// buildTranslationPrompt creates translation prompt
func (s *Service) buildTranslationPrompt(req *TranslationRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("Translate the following text from %s to %s:\n\n", req.FromLang, req.ToLang))

	if req.Context != "" {
		prompt.WriteString(fmt.Sprintf("Context: %s\n\n", req.Context))
	}

	prompt.WriteString(fmt.Sprintf("Text to translate: \"%s\"\n\n", req.Text))
	prompt.WriteString("Provide only the translated text without any additional formatting, quotes, or explanations.")

	return prompt.String()
}

// ValidateLanguageCode validates language code
func (s *Service) ValidateLanguageCode(langCode string) bool {
	// Simple validation - check that language code consists of 2-3 characters
	if len(langCode) < 2 || len(langCode) > 3 {
		return false
	}

	// Check that all characters are letters
	for _, char := range langCode {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') {
			return false
		}
	}

	return true
}
