# Configuration Guide

Руководство по настройке конфигурации AI Trader Bot System.

## 🎯 Обзор

AI Trader Bot использует TOML или YAML файлы для конфигурации. Все параметры можно настроить в соответствии с вашими требованиями к торговле и управлению рисками.

## 📁 Структура конфигурации

```toml
[bot]
name = "my-ai-trader"
version = "1.0.0"
environment = "production"

[logging]
level = "info"
format = "json"
audit_log_path = "logs/audit.log"
alert_log_path = "logs/alerts.log"

[trading_limits]
max_daily_volume = "10000factory/test/ndollar"
max_trades_per_day = 100
max_trades_per_hour = 10

[risk_management]
max_price_deviation = 0.05
stop_loss_percentage = 0.10
take_profit_percentage = 0.20

[oracle]
endpoint = "localhost:9090"
timeout = "10s"

[monitoring]
enabled = true
port = 8080
```

## ⚙️ Основные секции

### Bot Configuration

```toml
[bot]
name = "my-ai-trader"                    # Имя бота
version = "1.0.0"                        # Версия
environment = "production"               # Окружение (dev/test/prod)
description = "AI Trading Bot"           # Описание
```

### Logging Configuration

```toml
[logging]
level = "info"                           # Уровень логирования (debug/info/warn/error)
format = "json"                          # Формат логов (json/text)
audit_log_path = "logs/audit.log"       # Путь к файлу аудита
alert_log_path = "logs/alerts.log"      # Путь к файлу алертов
buffer_size = 1000                      # Размер буфера
flush_interval = "30s"                  # Интервал сброса буфера
max_file_size = 104857600              # Максимальный размер файла (100MB)
max_files = 10                          # Максимальное количество файлов
```

### Trading Limits

```toml
[trading_limits]
max_daily_volume = "10000factory/test/ndollar"  # Максимальный дневной объем
max_trades_per_day = 100                         # Максимальное количество сделок в день
max_trades_per_hour = 10                         # Максимальное количество сделок в час
max_single_trade = "1000factory/test/ndollar"    # Максимальная сумма одной сделки
allowed_symbols = ["BTC", "ETH", "USDC"]        # Разрешенные символы
```

### Risk Management

```toml
[risk_management]
max_price_deviation = 0.05              # Максимальное отклонение цены (5%)
stop_loss_percentage = 0.10             # Процент стоп-лосса (10%)
take_profit_percentage = 0.20           # Процент тейк-профита (20%)
cooldown_period = "5m"                  # Период охлаждения между сделками
max_consecutive_losses = 5              # Максимальное количество подряд идущих убытков
max_daily_loss = "5000factory/test/ndollar"  # Максимальный дневной убыток

[risk_management.emergency_stop]
enabled = true                          # Включить экстренную остановку
max_consecutive_losses = 5              # Максимум подряд идущих убытков
max_daily_loss = "10000factory/test/ndollar"  # Максимальный дневной убыток
max_oracle_downtime = "10m"             # Максимальное время недоступности oracle
```

### Oracle Configuration

```toml
[oracle]
endpoint = "localhost:9090"             # Адрес gRPC эндпоинта
timeout = "10s"                         # Таймаут запросов
retry_attempts = 3                      # Количество попыток повтора
retry_delay = "1s"                      # Задержка между попытками
price_sources = ["primary", "backup"]   # Источники цен
update_interval = "30s"                 # Интервал обновления цен
```

### Monitoring Configuration

```toml
[monitoring]
enabled = true                          # Включить мониторинг
port = 8080                             # Порт для API
metrics_enabled = true                  # Включить метрики
health_check_interval = "30s"          # Интервал проверки здоровья
alert_rules_enabled = true             # Включить правила алертов

[monitoring.notifiers]
console_enabled = true                  # Консольные уведомления
file_enabled = true                     # Файловые уведомления
webhook_enabled = false                 # Webhook уведомления
email_enabled = false                   # Email уведомления

[monitoring.webhook]
url = "https://hooks.slack.com/..."     # URL webhook
headers = { "Content-Type" = "application/json" }
timeout = "5s"

[monitoring.email]
smtp_host = "smtp.gmail.com"
smtp_port = 587
from = "alerts@your-domain.com"
to = ["admin@your-domain.com"]
```

## 🔧 Расширенные настройки

### Security Configuration

```toml
[security]
key_storage_type = "vault"              # Тип хранения ключей (vault/kms/file)
vault_endpoint = "https://vault.example.com"
vault_token_path = "/path/to/token"
kms_key_id = "arn:aws:kms:..."
encryption_enabled = true               # Включить шифрование
audit_enabled = true                    # Включить аудит
```

### Performance Configuration

```toml
[performance]
max_concurrent_trades = 5               # Максимальное количество одновременных сделок
trade_timeout = "30s"                   # Таймаут торговых операций
price_cache_ttl = "60s"                 # TTL кэша цен
connection_pool_size = 10               # Размер пула соединений
batch_size = 100                        # Размер батча для обработки
```

### Data Sources Configuration

```toml
[data_sources]
primary_oracle = "nuah-oracle"          # Основной oracle
backup_oracles = ["backup-oracle-1", "backup-oracle-2"]
external_apis = ["coinbase", "binance"] # Внешние API
data_validation_enabled = true         # Включить валидацию данных
cross_validation_threshold = 0.02      # Порог кросс-валидации (2%)
```

## 📊 Примеры конфигураций

### Консервативная конфигурация

```toml
[bot]
name = "conservative-trader"
environment = "production"

[trading_limits]
max_daily_volume = "5000factory/test/ndollar"
max_trades_per_day = 20
max_trades_per_hour = 2
max_single_trade = "500factory/test/ndollar"
allowed_symbols = ["BTC", "ETH"]

[risk_management]
max_price_deviation = 0.02              # 2%
stop_loss_percentage = 0.05             # 5%
take_profit_percentage = 0.10           # 10%
cooldown_period = "10m"                 # 10 минут
max_consecutive_losses = 3
max_daily_loss = "1000factory/test/ndollar"

[risk_management.emergency_stop]
enabled = true
max_consecutive_losses = 3
max_daily_loss = "2000factory/test/ndollar"
```

### Агрессивная конфигурация

```toml
[bot]
name = "aggressive-trader"
environment = "production"

[trading_limits]
max_daily_volume = "100000factory/test/ndollar"
max_trades_per_day = 1000
max_trades_per_hour = 100
max_single_trade = "10000factory/test/ndollar"
allowed_symbols = ["BTC", "ETH", "USDC", "ATOM", "OSMO"]

[risk_management]
max_price_deviation = 0.10              # 10%
stop_loss_percentage = 0.15              # 15%
take_profit_percentage = 0.30            # 30%
cooldown_period = "1m"                  # 1 минута
max_consecutive_losses = 10
max_daily_loss = "50000factory/test/ndollar"

[risk_management.emergency_stop]
enabled = true
max_consecutive_losses = 10
max_daily_loss = "100000factory/test/ndollar"
```

### Тестовая конфигурация

```toml
[bot]
name = "test-trader"
environment = "test"

[trading_limits]
max_daily_volume = "100factory/test/ndollar"
max_trades_per_day = 5
max_trades_per_hour = 1
max_single_trade = "10factory/test/ndollar"
allowed_symbols = ["BTC"]

[risk_management]
max_price_deviation = 0.50              # 50% для тестов
stop_loss_percentage = 0.20             # 20%
take_profit_percentage = 0.50           # 50%
cooldown_period = "30s"                 # 30 секунд
max_consecutive_losses = 2
max_daily_loss = "50factory/test/ndollar"

[logging]
level = "debug"                          # Подробное логирование
```

## 🔄 Динамическая конфигурация

### Hot Reload

```toml
[config]
hot_reload_enabled = true                # Включить горячую перезагрузку
reload_interval = "60s"                 # Интервал проверки изменений
config_file_path = "config.toml"        # Путь к файлу конфигурации
```

### Environment Variables

```bash
# Переопределение через переменные окружения
export AI_TRADER_LOG_LEVEL=debug
export AI_TRADER_ORACLE_ENDPOINT=localhost:9090
export AI_TRADER_MAX_DAILY_VOLUME=20000factory/test/ndollar
```

## 🧪 Валидация конфигурации

### Проверка конфигурации

```bash
# Валидация файла конфигурации
go run ./services/ai_trader/main.go --config config.toml --validate

# Проверка синтаксиса TOML
toml validate config.toml
```

### Тестирование конфигурации

```bash
# Запуск с тестовой конфигурацией
go run ./services/ai_trader/main.go --config test-config.toml --dry-run

# Проверка подключения к oracle
go run ./services/ai_trader/main.go --config config.toml --test-oracle
```

## 📋 Чек-лист конфигурации

### Обязательные параметры

- [ ] `bot.name` - Имя бота
- [ ] `trading_limits.max_daily_volume` - Максимальный дневной объем
- [ ] `trading_limits.max_trades_per_day` - Максимальное количество сделок
- [ ] `risk_management.stop_loss_percentage` - Процент стоп-лосса
- [ ] `oracle.endpoint` - Адрес oracle
- [ ] `monitoring.port` - Порт мониторинга

### Рекомендуемые параметры

- [ ] `risk_management.cooldown_period` - Период охлаждения
- [ ] `risk_management.max_consecutive_losses` - Максимум подряд идущих убытков
- [ ] `logging.level` - Уровень логирования
- [ ] `monitoring.enabled` - Включение мониторинга
- [ ] `security.encryption_enabled` - Шифрование

### Опциональные параметры

- [ ] `bot.description` - Описание бота
- [ ] `performance.max_concurrent_trades` - Максимум одновременных сделок
- [ ] `data_sources.backup_oracles` - Резервные oracle
- [ ] `monitoring.webhook` - Webhook уведомления

## 🆘 Устранение неполадок

### Частые ошибки конфигурации

1. **"invalid coin format"**:
   - Проверьте формат сумм: `"1000factory/test/ndollar"`
   - Убедитесь в правильности деноминации

2. **"invalid duration format"**:
   - Используйте правильный формат: `"5m"`, `"30s"`, `"1h"`
   - Проверьте синтаксис временных интервалов

3. **"oracle connection failed"**:
   - Проверьте доступность oracle эндпоинта
   - Убедитесь в правильности адреса и порта

4. **"invalid risk parameters"**:
   - Проверьте, что проценты указаны как десятичные дроби (0.05 = 5%)
   - Убедитесь в логичности значений (stop_loss < take_profit)

### Диагностические команды

```bash
# Проверка синтаксиса TOML
toml validate config.toml

# Проверка подключения к oracle
telnet oracle-host 9090

# Проверка доступности портов
netstat -tulpn | grep :8080

# Проверка прав доступа к файлам
ls -la config.toml logs/
```

## 📚 Дополнительные ресурсы

- [TOML Specification](https://toml.io/en/)
- [YAML Specification](https://yaml.org/spec/)
- [Go Configuration Best Practices](https://golang.org/doc/effective_go.html#configuration)
- [Production Configuration Patterns](https://12factor.net/config)

---

**⚠️ Важно**: Всегда тестируйте конфигурацию на тестовой среде перед применением в продакшене. Регулярно проверяйте и обновляйте параметры в соответствии с изменяющимися рыночными условиями.

