# Scripts Directory

Эта папка содержит все скрипты для управления Translation API.

## Структура

```
scripts/
├── deploy/           # Скрипты развертывания
│   ├── deploy.sh     # Полное развертывание в продакшене
│   └── update.sh     # Обновление приложения
├── ssl/              # Управление SSL сертификатами
│   ├── init-letsencrypt.sh  # Инициализация SSL
│   ├── renew-certs.sh       # Обновление сертификатов
│   └── cron-renew.sh        # Автоматическое обновление
└── health-check.sh   # Проверка здоровья системы
```

## Скрипты развертывания (`deploy/`)

### `deploy.sh`
Полное развертывание приложения в продакшене.

**Использование:**
```bash
make deploy
# или
./scripts/deploy/deploy.sh
```

**Что делает:**
- Проверяет предварительные требования
- Создает `.env.prod` если не существует
- Получает SSL сертификаты
- Собирает и запускает все сервисы
- Выполняет health check

### `update.sh`
Быстрое обновление приложения в продакшене.

**Использование:**
```bash
make update
# или
./scripts/deploy/update.sh
```

**Что делает:**
- Останавливает сервисы
- Обновляет код из git
- Пересобирает Docker образы
- Запускает обновленные сервисы
- Выполняет health check

## SSL скрипты (`ssl/`)

### `init-letsencrypt.sh`
Инициализация SSL сертификатов Let's Encrypt.

**Использование:**
```bash
make ssl-init
# или
./scripts/ssl/init-letsencrypt.sh
```

### `renew-certs.sh`
Обновление SSL сертификатов.

**Использование:**
```bash
make ssl-renew
# или
./scripts/ssl/renew-certs.sh
```

### `cron-renew.sh`
Скрипт для автоматического обновления через cron.

**Настройка cron:**
```bash
crontab -e
# Добавить строку:
0 */12 * * * /path/to/translation/scripts/ssl/cron-renew.sh
```

## Мониторинг

### `health-check.sh`
Комплексная проверка здоровья системы.

**Использование:**
```bash
make health-check
# или
./scripts/health-check.sh
```

**Проверяет:**
- Статус всех сервисов
- SSL сертификаты
- Сетевое подключение
- Health endpoints
- Использование ресурсов
- Логи

## Переменные окружения

Все скрипты используют переменные из `.env.prod`:

```bash
DOMAIN=your-domain.com
CERTBOT_EMAIL=your-email@example.com
REDIS_PASSWORD=your-secure-password
RABBITMQ_USER=translation_user
RABBITMQ_PASS=your-secure-password
OPENAI_API_KEY=your-openai-api-key
```

## Безопасность

- Все скрипты проверяют, что не запущены от root
- Проверяют наличие необходимых зависимостей
- Используют безопасные пароли
- Логируют все действия

## Troubleshooting

### Скрипт не выполняется
```bash
chmod +x scripts/*/script.sh
```

### Ошибки SSL
```bash
make ssl-init
make ssl-renew
```

### Проблемы с развертыванием
```bash
make health-check
make prod-logs
```

### Обновление не работает
```bash
git status
make update
``` 