package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "translation/docs"
	appTranslation "translation/internal/application/translation"
	"translation/internal/config"
	domainTranslation "translation/internal/domain/translation"
	"translation/internal/infrastructure/openai"
	"translation/internal/infrastructure/rabbitmq"
	redisRepo "translation/internal/infrastructure/redis"
	"translation/internal/interfaces/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/redis/go-redis/v9"
)

// @title Translation Service API
// @version 1.0
// @description A service for generating translations based on ARB files from Flutter applications using OpenAI API
// @termsOfService http://swagger.io/terms/

// @contact.name Kyrylo Kovalenko
// @contact.url https://kovalenko.tech
// @contact.email git@kovalenko.tech

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description API Key for authentication. Use format: Bearer YOUR_API_KEY

func main() {
	// Log application info
	log.Println("Translation Service - by Kyrylo Kovalenko (git@kovalenko.tech)")
	log.Println("Website: https://kovalenko.tech")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.URL,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Check Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize RabbitMQ service
	rabbitService, err := rabbitmq.NewService(cfg.RabbitMQ.URL, cfg.RabbitMQ.QueueName)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitService.Close()

	// Initialize OpenAI service
	openaiService := openai.NewService(cfg.OpenAI.APIKey)

	// Initialize repository
	repo := redisRepo.NewRepository(redisClient)

	// Initialize domain service
	domainService := domainTranslation.NewService(repo)

	// Initialize application service
	appService := appTranslation.NewService(domainService, openaiService, rabbitService)

	// Initialize HTTP handlers
	handler := http.NewHandler(appService)

	// Create Fiber application
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("HTTP Error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		},
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Setup routes
	http.SetupRoutes(app, handler, cfg.Server.APIKey)

	// Recover incomplete requests on startup
	log.Println("Recovering incomplete translation requests...")
	if err := appService.RecoverIncompleteRequests(context.Background()); err != nil {
		log.Printf("Failed to recover incomplete requests: %v", err)
	}

	// Start consumer in a separate goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := appService.StartConsumer(ctx); err != nil {
			log.Printf("Failed to start consumer: %v", err)
		}
	}()

	// Configure graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		cancel() // Stop consumer
		if err := app.Shutdown(); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
