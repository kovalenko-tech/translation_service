# Quick Start - Production Deployment

**Author:** [Kyrylo Kovalenko](https://kovalenko.tech) - git@kovalenko.tech

## Quick Start (5 minutes)

### 1. Preparation
```bash
# Clone the repository
git clone <your-repo>
cd translation

# Create configuration file
cp env.prod.example .env.prod
```

### 2. Configuration
Edit `.env.prod`:
```bash
DOMAIN=your-domain.com
CERTBOT_EMAIL=your-email@example.com
REDIS_PASSWORD=your-secure-password
RABBITMQ_USER=translation_user
RABBITMQ_PASS=your-secure-password
OPENAI_API_KEY=your-openai-api-key
```

### 3. Deployment
```bash
# Automatic deployment
make deploy
```

Or manually:
```bash
# Get SSL certificates
make ssl-init

# Start production
make prod-up

# Health check
make health-check
```

Or using the deployment script:
```bash
make deploy
```

### 4. Verification
```bash
# Service status
make prod-status

# Logs
make prod-logs

# Health check
make health-check
```

## Application Access

- **API**: https://your-domain.com/api/
- **Swagger**: https://your-domain.com/docs/
- **Health Check**: https://your-domain.com/api/health

## Management

```bash
# Stop
make prod-down

# Restart
make prod-down && make prod-up

# SSL renewal
make ssl-renew

# Application update
make update
```

## Automatic SSL Renewal

Add to crontab:
```bash
crontab -e
# Add line:
0 */12 * * * /path/to/translation/scripts/ssl/cron-renew.sh
```

## Troubleshooting

### SSL certificates not working
```bash
make ssl-init
```

### Services not starting
```bash
make health-check
make prod-logs
```

### Application update
```bash
make update
```

## Security

- ✅ HTTPS with modern protocols
- ✅ Rate limiting (10 req/sec)
- ✅ Security headers
- ✅ Isolated Docker network
- ✅ Secure passwords for Redis/RabbitMQ

## Monitoring

- Health check: `make health-check`
- Logs: `make prod-logs`
- Status: `make prod-status`
- Resources: `docker stats` 