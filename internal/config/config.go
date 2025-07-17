// Translation Service Configuration
// Author: Kyrylo Kovalenko (git@kovalenko.tech)
// Website: https://kovalenko.tech
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config represents application configuration
type Config struct {
	Server   ServerConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	OpenAI   OpenAIConfig
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port   string
	Host   string
	APIKey string
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// RabbitMQConfig represents RabbitMQ configuration
type RabbitMQConfig struct {
	URL       string
	QueueName string
}

// OpenAIConfig represents OpenAI configuration
type OpenAIConfig struct {
	APIKey string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:   getEnv("SERVER_PORT", "8080"),
			Host:   getEnv("SERVER_HOST", "0.0.0.0"),
			APIKey: getEnv("API_KEY", ""),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RabbitMQ: RabbitMQConfig{
			URL:       getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			QueueName: getEnv("RABBITMQ_QUEUE", "translation_tasks"),
		},
		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPENAI_API_KEY", ""),
		},
	}

	// Validate required parameters
	if config.OpenAI.APIKey == "" {
		return nil, &ConfigError{Message: "OPENAI_API_KEY is required"}
	}

	if config.Server.APIKey == "" {
		return nil, &ConfigError{Message: "API_KEY is required"}
	}

	return config, nil
}

// getEnv gets environment variable value or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable value as int or returns default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// ConfigError represents configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
