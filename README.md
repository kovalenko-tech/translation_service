# Translation Service

A service for generating translations based on ARB files from Flutter applications using OpenAI API.

**Author:** [Kyrylo Kovalenko](https://kovalenko.tech) - git@kovalenko.tech

## Architecture

The project is built using Domain-Driven Design (DDD) architecture:

- **Domain Layer** - domain logic and business rules
- **Application Layer** - coordination of domain services
- **Infrastructure Layer** - external services (Redis, RabbitMQ, OpenAI)
- **Interface Layer** - HTTP API

## Features

- Processing ARB files from Flutter applications
- Key filtering (excluding keys starting with @ or @@)
- Asynchronous translation processing via RabbitMQ
- Caching in Redis
- Translation generation via OpenAI API
- REST API for creating requests and getting status
- Translation key management (create, read, delete)
- **Direct translation caching** - cache translations without running translation process
- **Smart translation skipping** - skip translation if all required translations already exist
- Interactive API documentation with Swagger
- Real-time translation status tracking
- Multi-language translation support

## API Endpoints

### POST /api/v1/translations
Creates a new translation request.

**Request Body:**
```json
{
  "source_data": {
    "hello": "Hello World",
    "welcome": "Welcome to our app"
  },
  "languages": ["es", "fr", "de"]
}
```

**Response:**
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "message": "Translation request created successfully and queued for processing"
}
```

### GET /api/v1/translations/:id
Gets the status and results of a translation request.

**Response:**
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "source_data": {
    "hello": "Hello World",
    "welcome": "Welcome to our app"
  },
  "languages": ["es", "fr", "de"],
  "translated_data": {
    "es": {
      "hello": "Hola Mundo",
      "welcome": "Bienvenido a nuestra aplicación"
    },
    "fr": {
      "hello": "Bonjour le monde",
      "welcome": "Bienvenue dans notre application"
    },
    "de": {
      "hello": "Hallo Welt",
      "welcome": "Willkommen in unserer App"
    }
  },
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:05:00Z",
  "completed_at": "2024-01-01T12:05:00Z"
}
```

**Note:** The `translated_data` field is only included when the request status is `completed`.

### POST /api/v1/translations/cache
Caches translations for keys without running translation process. English translations are required for all keys.

**Request Body:**
```json
{
  "translations": {
    "en": {
      "hello": "Hello World",
      "welcome": "Welcome to our app"
    },
    "es": {
      "hello": "Hola Mundo",
      "welcome": "Bienvenido a nuestra aplicación"
    },
    "fr": {
      "hello": "Bonjour le monde",
      "welcome": "Bienvenue dans notre application"
    }
  }
}
```

**Success Response (200):**
```json
{
  "message": "Translations cached successfully",
  "count": 2
}
```

**Partial Success Response (207 Multi-Status):**
```json
{
  "error": "Some translations could not be cached - English translations are required for all keys",
  "skipped_keys": ["missingKey"],
  "success_count": 1,
  "total_keys": 2
}
```

**Note:** 
- English translations (`en`) are mandatory for all keys
- Keys without English translations will be skipped
- Returns 207 status when some keys are skipped
- Returns 200 status when all keys are successfully cached

### DELETE /api/v1/translations/:key
Deletes a translation key and all its translations.

**Response:** `204 No Content`

### GET /api/v1/health
Service health check.

**Response:**
```json
{
  "status": "ok",
  "message": "Translation service is running",
  "author": "Kyrylo Kovalenko",
  "contact": "git@kovalenko.tech",
  "website": "https://kovalenko.tech"
}
```

## API Security

All API endpoints (except `/api/v1/health`) are protected with an API key. To access protected endpoints, you need to pass the token in the `Authorization` header.

### Generating API Key

To generate a secure API key, use the command:
```bash
make generate-api-key
```

### Using API Key

Add the generated key to your `.env` file:
```bash
API_KEY=your_generated_api_key_here
```

### Example requests with API key

**Create translation request:**
```bash
curl -X POST http://localhost:8080/api/v1/translations \
  -H "Authorization: Bearer your_api_key_here" \
  -H "Content-Type: application/json" \
  -d '{
    "source_data": {
      "hello": "Hello World",
      "welcome": "Welcome to our app"
    },
    "languages": ["es", "fr"]
  }'
```

**Get translation status:**
```bash
curl -X GET http://localhost:8080/api/v1/translations/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer your_api_key_here"
```

**Delete translation key:**
```bash
curl -X DELETE http://localhost:8080/api/v1/translations/hello \
  -H "Authorization: Bearer your_api_key_here"
```

**Cache translations:**
```bash
curl -X POST http://localhost:8080/api/v1/translations/cache \
  -H "Authorization: Bearer your_api_key_here" \
  -H "Content-Type: application/json" \
  -d '{
    "translations": {
      "en": {
        "hello": "Hello World",
        "welcome": "Welcome to our app"
      },
      "es": {
        "hello": "Hola Mundo",
        "welcome": "Bienvenido a nuestra aplicación"
      },
      "fr": {
        "hello": "Bonjour le monde",
        "welcome": "Bienvenue dans notre application"
      }
    }
  }'
```

**Important:** Keep the API key in a secure place and do not share it in public repositories.

## Installation and Setup

### Prerequisites

- Go 1.23+
- Redis
- RabbitMQ
- OpenAI API key

### Quick Start Options

- **[Development Setup](#development-setup)** - Local development environment
- **[Production Deployment](QUICK_START_PROD.md)** - Quick production deployment (5 minutes)
- **[Production Guide](PRODUCTION.md)** - Detailed production deployment guide

### Development Setup

#### 1. Clone the repository

```bash
git clone <repository-url>
cd translation
```

#### 2. Install dependencies

```bash
go mod download
```

#### 3. Configure settings

Copy the configuration file:
```bash
cp env.example .env
```

Edit the `.env` file with your settings:
```bash
# OpenAI API key (required)
OPENAI_API_KEY=your_openai_api_key_here

# Redis settings (default)
REDIS_URL=redis://localhost:6379

# RabbitMQ settings (default)
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

#### 4. Start dependencies

##### Redis
```bash
# Docker
docker run -d -p 6379:6379 redis:alpine

# Or locally
redis-server
```

##### RabbitMQ
```bash
# Docker
docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:management

# Or locally
rabbitmq-server
```

#### 5. Start the service

```bash
go run cmd/server/main.go
```

The service will be available at `http://localhost:8080`

## Project Structure

```
translation/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── domain/
│   │   └── translation/
│   │       ├── entity.go           # Domain entities
│   │       ├── repository.go       # Repository interface
│   │       └── service.go          # Domain service
│   ├── application/
│   │   └── translation/
│   │       └── service.go          # Application service
│   ├── infrastructure/
│   │   ├── redis/
│   │   │   └── repository.go       # Redis repository
│   │   ├── rabbitmq/
│   │   │   └── service.go          # RabbitMQ service
│   │   └── openai/
│   │       └── service.go          # OpenAI service
│   ├── interfaces/
│   │   └── http/
│   │       └── handlers.go         # HTTP handlers
│   └── config/
│       └── config.go               # Configuration
├── go.mod
├── go.sum
├── env.example
└── README.md
```

## ARB File Processing

The service automatically:
1. Extracts all keys from ARB file JSON data
2. Excludes keys starting with `@` or `@@`
3. Saves unique keys in Redis
4. Generates translations for new keys via OpenAI
5. Updates existing translations when necessary

## Request Statuses

- `pending` - request created and waiting for processing
- `processing` - request is being processed
- `completed` - request successfully completed
- `failed` - error occurred during processing

## Logging

The service outputs detailed logs to stdout, including:
- Request creation and processing
- Translation status
- External service connection errors
- HTTP requests

## API Documentation

The service provides interactive API documentation using Swagger:

- **Swagger UI**: http://localhost:8080/docs/
- **OpenAPI JSON**: http://localhost:8080/docs/doc.json

## Monitoring

For monitoring queue status, you can use:
- RabbitMQ Management UI (http://localhost:15672)
- Redis CLI for viewing cached data
- Health check endpoint `/api/v1/health`

## Translation Caching

The service supports direct translation caching to improve performance and reduce API costs:

### Benefits
- **Performance**: Instant access to cached translations
- **Cost reduction**: Avoid unnecessary OpenAI API calls
- **Flexibility**: Cache translations from external sources
- **Smart processing**: Skip translation if all required translations exist

### Usage Scenarios
1. **Pre-loading translations**: Cache existing translations before processing new requests
2. **External translation sources**: Import translations from other systems
3. **Manual corrections**: Cache corrected translations without re-translation
4. **Batch operations**: Cache multiple translations at once

### Caching Rules
- English translations (`en`) are mandatory for all keys
- Keys without English translations are skipped
- Existing translations are updated if provided
- The service automatically skips translation requests when all required translations exist in cache

## Production Deployment

For production deployment, we provide comprehensive documentation:

### Quick Start (5 minutes)
See [QUICK_START_PROD.md](QUICK_START_PROD.md) for a fast production deployment guide.

### Detailed Production Guide
See [PRODUCTION.md](PRODUCTION.md) for a complete production deployment guide including:
- Architecture overview
- Security configuration
- SSL/TLS setup
- Scaling strategies
- Monitoring and alerts
- Backup procedures
- Troubleshooting 