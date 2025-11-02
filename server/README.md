# Server

HTTP сервер для блокчейна.

📖 **Документация**: См. [docs/](./docs/) для подробной документации на английском языке.

## Documentation

- [Authentication System](./docs/AUTHENTICATION.md) - Complete authentication API documentation

## Запуск

### Автоматический запуск (рекомендуется)

С `air` сервер автоматически запустит PostgreSQL через Docker перед стартом:

```bash
air
```

`air` будет:
- Автоматически запускать PostgreSQL через Docker при каждом перезапуске
- Ждать готовности БД перед запуском сервера
- Автоматически перезапускать сервер при изменении файлов в директории `server/`

### Ручной запуск

#### 1. Запуск PostgreSQL через Docker

```bash
cd server
docker-compose up -d
```

Проверить статус:
```bash
docker-compose ps
```

#### 2. Запуск сервера

```bash
go run ./server
```

## Endpoints

### Health Check (Server)
```bash
curl http://localhost:8080/health
```

Ответ:
```json
{
  "status": "ok",
  "timestamp": "2025-11-02T22:35:27.848218+07:00"
}
```

### Health Check (Database)
```bash
curl http://localhost:8080/health/db
```

Ответ (если БД доступна):
```json
{
  "status": "ok",
  "timestamp": "2025-11-02T22:35:27.848218+07:00"
}
```

Ответ (если БД недоступна):
```json
{
  "status": "error",
  "timestamp": "2025-11-02T22:35:27.848218+07:00",
  "error": "database health check failed: ..."
}
```

## Конфигурация

Настройки можно изменить через переменные окружения:

### Сервер
- `SERVER_ADDRESS` - адрес сервера (по умолчанию: `0.0.0.0:8080`)

### База данных
- `DB_HOST` - хост БД (по умолчанию: `localhost`)
- `DB_PORT` - порт БД (по умолчанию: `5432`)
- `DB_USER` - пользователь БД (по умолчанию: `postgres`)
- `DB_PASSWORD` - пароль БД (по умолчанию: `postgres`)
- `DB_NAME` - имя БД (по умолчанию: `serverdb`)
- `DB_SSLMODE` - режим SSL (по умолчанию: `disable`)

### Аутентификация
- `JWT_SECRET` - секретный ключ для JWT токенов (по умолчанию: `your-secret-key-change-in-production`)
- `AUTH_MASTER_KEY` - мастер-ключ для шифрования приватных ключей кошельков (обязательно!)

**Важно:** Для продакшена обязательно установите `AUTH_MASTER_KEY` и `JWT_SECRET`!
```bash
export AUTH_MASTER_KEY="$(openssl rand -base64 32)"
export JWT_SECRET="$(openssl rand -base64 32)"
```

### Логирование
- `LOG_ENABLED` - включить/выключить логирование (по умолчанию: `true`)
- `LOG_LEVEL` - уровень логирования: `debug`, `info`, `warn`, `error` (по умолчанию: `debug`)
- `LOG_FORMAT` - формат логов: `json`, `text` (по умолчанию: `text`)
- `ENVIRONMENT` - окружение: `dev`, `prod` (по умолчанию: `dev`)

#### Примеры использования

**Для разработки (цветной вывод, подробные логи):**
```bash
ENVIRONMENT=dev LOG_LEVEL=debug LOG_FORMAT=text air
```

**Для продакшена (JSON формат, минимальные логи, можно отключить):**
```bash
ENVIRONMENT=prod LOG_LEVEL=info LOG_FORMAT=json LOG_ENABLED=true air
# или полностью отключить логирование:
ENVIRONMENT=prod LOG_ENABLED=false air
```

## API Endpoints

### Authentication

#### POST /api/auth/register
Регистрация нового пользователя через веб (email/password).

**Request:**
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "username",
    "created_at": "2025-11-02T23:00:00Z"
  },
  "wallet": {
    "id": 1,
    "user_id": 1,
    "address": "cosmos1...",
    "created_at": "2025-11-02T23:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Registration successful"
}
```

#### POST /api/auth/login
Вход пользователя через веб (email/password).

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "user": { ... },
  "wallet": { ... },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login successful"
}
```

#### POST /api/auth/telegram
Регистрация/вход через Telegram WebApp.

**Request:**
```json
{
  "id": 123456789,
  "first_name": "John",
  "last_name": "Doe",
  "username": "johndoe",
  "auth_date": 1698950000,
  "hash": "telegram_hash_here"
}
```

**Response:**
```json
{
  "user": { ... },
  "wallet": { ... },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Telegram authentication successful"
}
```

#### GET /api/auth/me
Получение информации о текущем пользователе (требует токен в заголовке `Authorization: Bearer <token>`).

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "username": "username",
  "created_at": "2025-11-02T23:00:00Z"
}
```

## Структура

```
server/
├── api/              # REST API хендлеры
│   ├── auth.go       # Аутентификация эндпоинты
│   ├── db_health.go  # Health check БД
│   ├── health.go     # Health check сервера
│   └── router.go     # Роутер
├── auth/             # Система аутентификации
│   ├── crypto.go     # Шифрование ключей
│   ├── jwt.go        # JWT токены
│   ├── middleware.go # Middleware аутентификации
│   ├── models.go     # Модели данных
│   ├── repository.go # Репозиторий БД
│   ├── service.go    # Бизнес-логика
│   └── wallet.go     # Генерация кошельков
├── config/           # Конфигурация
├── database/          # Модуль работы с БД
│   ├── database.go   # Подключение к БД
│   └── migrations.go # Миграции
├── docs/              # Документация (на английском)
│   ├── README.md     # Индекс документации
│   └── AUTHENTICATION.md # Документация аутентификации
├── logger/            # Система логирования
├── migrations/        # SQL миграции
│   └── 001_initial_auth.sql
├── scripts/           # Вспомогательные скрипты
│   ├── start_with_auth.sh  # Запуск с ключами
│   ├── test_auth.sh        # Тестирование API
│   └── wait-for-db.sh      # Ожидание БД
├── docker-compose.yml # PostgreSQL конфигурация
└── main.go           # Точка входа
```

## Логирование

Сервер использует структурированное логирование с поддержкой:
- **HTTP middleware** - автоматическое логирование всех запросов
- **Разные форматы** - JSON для продакшена, цветной текст для разработки
- **Уровни логирования** - debug, info, warn, error
- **Включение/отключение** - можно полностью отключить через `LOG_ENABLED=false`

### Примеры логов

**В режиме разработки (dev):**
```
2025-11-02 23:00:00 Server initializing...
2025-11-02 23:00:01 Database connected successfully
2025-11-02 23:00:01 Server starting address=0.0.0.0:8080
2025-11-02 23:00:05 HTTP request method=GET path=/health status=200 duration=2 remote_ip=127.0.0.1
```

**В режиме продакшена (prod, JSON):**
```json
{"level":"info","msg":"Server initializing...","time":"2025-11-02T23:00:00Z"}
{"level":"info","msg":"Database connected successfully","time":"2025-11-02T23:00:01Z"}
{"level":"info","msg":"HTTP request","method":"GET","path":"/health","status":200,"duration":2,"time":"2025-11-02T23:00:05Z"}
```
