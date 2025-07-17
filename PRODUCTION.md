# Production Deployment Guide

**Author:** [Kyrylo Kovalenko](https://kovalenko.tech) - git@kovalenko.tech

This document describes the process of deploying the Translation API in production using Docker Compose, Nginx reverse proxy, and SSL certificates.

## Architecture

```
Internet → Nginx (SSL termination) → Load Balancer → Multiple App Instances
                                    ↓
                              Redis + RabbitMQ
```

### Components

- **Nginx**: Reverse proxy with SSL termination and load balancing
- **Certbot**: Automatic SSL certificate acquisition and renewal
- **Application**: 3 application instances for high availability
- **Redis**: Caching and sessions
- **RabbitMQ**: Message queues

## Prerequisites

1. Server with Docker and Docker Compose
2. Domain pointing to the server IP
3. Open ports 80 and 443

## Setup

### 1. Cloning and Configuration

```bash
git clone <your-repo>
cd translation
```

### 2. Environment Configuration

```bash
cp env.prod.example .env.prod
```

Edit `.env.prod`:

```bash
# Replace with your domain
DOMAIN=your-domain.com

# Email for SSL certificates
CERTBOT_EMAIL=your-email@example.com

# Secure passwords
REDIS_PASSWORD=your-secure-redis-password
RABBITMQ_USER=translation_user
RABBITMQ_PASS=your-secure-rabbitmq-password

# OpenAI API key
OPENAI_API_KEY=your-openai-api-key
```

### 3. Nginx Configuration

Edit `nginx/conf.d/default.conf`:

```bash
# Replace your-domain.com with your domain
server_name your-domain.com www.your-domain.com;
```

### 4. SSL Certificate Acquisition

```bash
make ssl-init
```

### 5. Production Startup

```bash
# Automatic deployment
make deploy

# Or manually
make prod-up
```

## Management

### Main Commands

```bash
# Automatic deployment
make deploy

# Start production
make prod-up

# Stop production
make prod-down

# View logs
make prod-logs

# Service status
make prod-status

# SSL certificate renewal
make ssl-renew

# Health check
make health-check
```

### Monitoring

```bash
# Check status of all services
docker-compose -f docker-compose.prod.yml ps

# Nginx logs
docker-compose -f docker-compose.prod.yml logs nginx

# Application logs
docker-compose -f docker-compose.prod.yml logs app

# Redis logs
docker-compose -f docker-compose.prod.yml logs redis

# RabbitMQ logs
docker-compose -f docker-compose.prod.yml logs rabbitmq
```

### Application Updates

```bash
# Automatic update
make update

# Or manually
make prod-down
make prod-build
make prod-up
```

## Security

### SSL/TLS
- Automatic certificate acquisition through Let's Encrypt
- Forced HTTP → HTTPS redirect
- Modern SSL protocols (TLS 1.2, 1.3)
- Secure ciphers

### Security Headers
- HSTS (HTTP Strict Transport Security)
- X-Frame-Options
- X-Content-Type-Options
- X-XSS-Protection
- Referrer-Policy

### Rate Limiting
- Request limit: 10 requests/sec on API
- Burst: up to 20 requests

### Network Isolation
- All services in separate Docker network
- External ports only for nginx (80, 443)

## Scaling

### Horizontal Scaling

To increase the number of application instances:

```bash
# In docker-compose.prod.yml change:
deploy:
  replicas: 5  # Instead of 3
```

### Vertical Scaling

Add resource limits in `docker-compose.prod.yml`:

```yaml
app:
  # ... existing config ...
  deploy:
    resources:
      limits:
        cpus: '1.0'
        memory: 1G
      reservations:
        cpus: '0.5'
        memory: 512M
```

## Backup

### Redis
```bash
# Create backup
docker exec translation-redis redis-cli BGSAVE

# Copy file
docker cp translation-redis:/data/dump.rdb ./backup/redis-$(date +%Y%m%d).rdb
```

### RabbitMQ
```bash
# Export configuration
docker exec translation-rabbitmq rabbitmqctl export_definitions > ./backup/rabbitmq-$(date +%Y%m%d).json
```

## Troubleshooting

### SSL certificates not updating
```bash
# Check certificate status
docker-compose -f docker-compose.prod.yml run --rm certbot certificates

# Force renewal
make ssl-renew
```

### Nginx not starting
```bash
# Check configuration
docker-compose -f docker-compose.prod.yml exec nginx nginx -t

# View logs
docker-compose -f docker-compose.prod.yml logs nginx
```

### Application unavailable
```bash
# Health check
curl -k https://your-domain.com/api/health

# Check application logs
docker-compose -f docker-compose.prod.yml logs app
```

## Automatic SSL Renewal

Add to crontab for automatic certificate renewal:

```bash
# Renewal every 12 hours
0 */12 * * * cd /path/to/translation && make ssl-renew
```

## Monitoring and Alerts

It's recommended to set up monitoring:

- **Health checks**: `/api/health`
- **Metrics**: Prometheus + Grafana
- **Logs**: ELK Stack or Fluentd
- **Alerts**: Slack/Email notifications

## Performance

### Nginx Optimization
- Gzip compression enabled
- Static file caching
- Keep-alive connections

### Application Optimization
- 3 instances for load balancing
- Connection pooling for Redis/RabbitMQ
- Graceful shutdown

### Performance Monitoring
```bash
# Resource usage
docker stats

# Network connections
docker-compose -f docker-compose.prod.yml exec nginx netstat -tulpn
``` 