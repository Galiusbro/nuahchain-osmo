# AI Trader Bot System

Автоматизированная система торговли с делегированием полномочий для блокчейна Nuah Chain.

## 🎯 Обзор

AI Trader Bot System позволяет пользователям делегировать торговые полномочия AI-агенту, который может автономно выполнять торговые операции через модуль `x/assets`, не требуя постоянного участия пользователя.

### Ключевые возможности

- **Делегирование полномочий**: Один раз настроить разрешения для AI-агента
- **Автономная торговля**: AI-агент принимает решения и выполняет сделки
- **Управление рисками**: Встроенная система контроля лимитов и политик
- **Мониторинг и аудит**: Полная прозрачность всех операций
- **Гибкая конфигурация**: Настройка под конкретные потребности

## 🏗️ Архитектура

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Wallet   │    │  AI Trader Bot  │    │   Nuah Chain    │
│                 │    │                 │    │                 │
│ 1. Grant Authz  │───▶│ 2. Load Config  │───▶│ 3. Execute Txs  │
│ 2. Grant Fee    │    │ 3. Risk Check   │    │ 4. Query Oracle │
│ 3. Monitor      │◀───│ 4. Trade Logic  │◀───│ 5. Update State │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Компоненты системы

1. **Authz & Feegrant**: Делегирование полномочий и оплата комиссий
2. **Configuration**: Настройка торговых параметров и лимитов
3. **Oracle Client**: Получение ценовых данных
4. **Trading Client**: Выполнение торговых операций
5. **Risk Management**: Контроль рисков и соблюдение лимитов
6. **Monitoring**: Логирование и мониторинг операций

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.21+
- Nuah Chain node (локальный или тестовая сеть)
- Ключи для пользователя и AI-агента

### Установка

```bash
# Клонировать репозиторий
git clone <repository-url>
cd nuahchain_osmosis

# Установить зависимости
go mod download

# Собрать проект
make build
```

### Настройка полномочий

1. **Выдача торговых полномочий**:
```bash
# Запустить интерактивный скрипт
./scripts/ai_trader_grant.sh

# Или выполнить команды вручную
nuahd tx authz grant <ai-agent-address> generic --msg-type="/nuah.assets.MsgBuyAsset" --from <user-key> --chain-id nuah-testnet
nuahd tx authz grant <ai-agent-address> generic --msg-type="/nuah.assets.MsgSellAsset" --from <user-key> --chain-id nuah-testnet
```

2. **Выдача полномочий на оплату комиссий** (опционально):
```bash
nuahd tx feegrant grant <ai-agent-address> <user-address> --spend-limit 1000factory/test/ndollar --from <user-key> --chain-id nuah-testnet
```

### Конфигурация бота

1. **Создать конфигурационный файл**:
```bash
cp services/ai_trader/config/example.toml my_bot_config.toml
```

2. **Настроить параметры**:
```toml
[bot]
name = "my-ai-trader"
version = "1.0.0"

[trading_limits]
max_daily_volume = "10000factory/test/ndollar"
max_trades_per_day = 100
max_trades_per_hour = 10
max_single_trade = "1000factory/test/ndollar"

[risk_management]
max_price_deviation = 0.05
stop_loss_percentage = 0.10
take_profit_percentage = 0.20
cooldown_period = "5m"
max_consecutive_losses = 5
max_daily_loss = "5000factory/test/ndollar"

[oracle]
endpoint = "localhost:9090"
timeout = "10s"
retry_attempts = 3
```

### Запуск бота

```bash
# Запустить AI trader bot
go run ./services/ai_trader/main.go --config my_bot_config.toml

# Или запустить мониторинг отдельно
go run ./services/ai_trader/monitoring/example/main.go
```

## 📋 Подробные руководства

- [**Production Setup Guide**](docs/PRODUCTION_SETUP_GUIDE.md) - Полное руководство по развертыванию в продакшене
- [**Authz & Feegrant Guide**](docs/AUTHZ_FEEGRANT_GUIDE.md) - Настройка делегирования полномочий
- [**Configuration Guide**](docs/CONFIGURATION_GUIDE.md) - Детальная настройка параметров
- [**Risk Management Guide**](docs/RISK_MANAGEMENT_GUIDE.md) - Управление рисками
- [**Monitoring Guide**](docs/MONITORING_GUIDE.md) - Мониторинг и аудит
- [**Troubleshooting Guide**](docs/TROUBLESHOOTING_GUIDE.md) - Решение проблем

## 🔧 API Документация

### REST API (Мониторинг)

- `GET /health` - Проверка состояния
- `GET /api/v1/events` - Получение событий аудита
- `GET /api/v1/alerts` - Получение алертов
- `GET /api/v1/history/{trader}` - История торгов
- `GET /api/v1/metrics` - Системные метрики

### gRPC API

- `QueryPrice` - Получение ценовых данных
- `MsgBuyAsset` - Покупка актива
- `MsgSellAsset` - Продажа актива
- `MsgExec` - Выполнение делегированных операций

## 🛡️ Безопасность

### Рекомендации по безопасности

1. **Хранение ключей**:
   - Используйте аппаратные кошельки для пользовательских ключей
   - Храните ключи AI-агента в защищенном хранилище (KMS, Vault)
   - Никогда не коммитьте ключи в репозиторий

2. **Сетевая безопасность**:
   - Используйте TLS для всех соединений
   - Ограничьте доступ к API эндпоинтам
   - Настройте файрвол для блокчейн ноды

3. **Мониторинг**:
   - Настройте алерты на подозрительную активность
   - Регулярно проверяйте логи
   - Мониторьте использование лимитов

### Отзыв полномочий

В случае компрометации или необходимости остановки:

```bash
# Отозвать торговые полномочия
nuahd tx authz revoke <ai-agent-address> /nuah.assets.MsgBuyAsset --from <user-key>
nuahd tx authz revoke <ai-agent-address> /nuah.assets.MsgSellAsset --from <user-key>

# Отозвать полномочия на оплату комиссий
nuahd tx feegrant revoke <ai-agent-address> <user-address> --from <user-key>
```

## 📊 Мониторинг и метрики

### Ключевые метрики

- **Торговые метрики**: Объем торгов, количество сделок, прибыль/убыток
- **Рисковые метрики**: Нарушения лимитов, срабатывания стоп-лоссов
- **Системные метрики**: Время отклика, доступность oracle, ошибки

### Алерты

Система автоматически генерирует алерты при:
- Превышении лимитов
- Нарушении политик
- Срабатывании emergency stop
- Системных ошибках

## 🧪 Тестирование

### Запуск тестов

```bash
# Все тесты
go test ./services/ai_trader/... -v

# Конкретные компоненты
go test ./services/ai_trader/config/... -v
go test ./services/ai_trader/client/... -v
go test ./services/ai_trader/risk/... -v
go test ./services/ai_trader/monitoring/... -v

# Интеграционные тесты
go test ./services/ai_trader/... -v -tags=integration
```

### Тестовые сети

```bash
# Запуск локальной тестовой сети
make local-testnet

# Проверка состояния
nuahd status --node tcp://localhost:26657
```

## 🤝 Вклад в проект

### Разработка

1. Форкните репозиторий
2. Создайте feature branch
3. Внесите изменения
4. Добавьте тесты
5. Создайте pull request

### Стандарты кода

- Используйте `gofmt` для форматирования
- Запускайте `golint` для проверки стиля
- Покрывайте код тестами (минимум 80%)
- Документируйте публичные API

## 📝 Лицензия

Проект распространяется под лицензией [MIT License](LICENSE).

## 🆘 Поддержка

- **Документация**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/nuahchain_osmosis/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/nuahchain_osmosis/discussions)

## 📈 Roadmap

- [ ] Интеграция с дополнительными DEX
- [ ] Машинное обучение для оптимизации стратегий
- [ ] Веб-интерфейс для управления
- [ ] Мобильное приложение для мониторинга
- [ ] Интеграция с внешними источниками данных

---

**⚠️ Важно**: Данная система предназначена для тестирования и разработки. Для использования в продакшене требуется дополнительная настройка безопасности и аудит.
