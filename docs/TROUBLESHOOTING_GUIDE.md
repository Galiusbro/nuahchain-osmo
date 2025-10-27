# Troubleshooting Guide

Руководство по устранению неполадок AI Trader Bot System.

## 🎯 Обзор

Данное руководство поможет диагностировать и решить наиболее распространенные проблемы при работе с AI Trader Bot System.

## 🔍 Диагностика проблем

### Системная диагностика

```bash
# Проверка состояния системы
systemctl status ai-trader.service
systemctl status ai-trader-monitoring.service

# Проверка логов
journalctl -u ai-trader.service -f
journalctl -u ai-trader-monitoring.service -f

# Проверка ресурсов
htop
df -h
free -h
```

### Проверка конфигурации

```bash
# Валидация конфигурации
go run ./services/ai_trader/main.go --config config.toml --validate

# Проверка синтаксиса TOML
toml validate config.toml

# Проверка переменных окружения
env | grep AI_TRADER
```

## 🚨 Частые проблемы

### 1. Проблемы с запуском

#### Сервис не запускается

**Симптомы:**
- `systemctl start ai-trader.service` завершается с ошибкой
- В логах: `Failed to start AI Trader Bot Service`

**Причины и решения:**

1. **Неправильная конфигурация**:
   ```bash
   # Проверка конфигурации
   go run ./services/ai_trader/main.go --config config.toml --validate

   # Исправление ошибок в config.toml
   ```

2. **Недостаточно прав**:
   ```bash
   # Проверка прав доступа
   ls -la /home/aitrader/ai-trader-bot/

   # Исправление прав
   sudo chown -R aitrader:aitrader /home/aitrader/ai-trader-bot/
   sudo chmod +x /home/aitrader/ai-trader-bot/bin/ai-trader
   ```

3. **Порт занят**:
   ```bash
   # Проверка занятых портов
   netstat -tulpn | grep :8080

   # Освобождение порта
   sudo fuser -k 8080/tcp
   ```

#### Ошибки подключения к блокчейну

**Симптомы:**
- `Failed to connect to oracle endpoint`
- `gRPC connection failed`

**Решения:**

1. **Проверка доступности ноды**:
   ```bash
   # Проверка подключения
   telnet your-nuah-node 9090

   # Проверка статуса ноды
   nuahd status --node tcp://your-nuah-node:26657
   ```

2. **Проверка конфигурации oracle**:
   ```toml
   [oracle]
   endpoint = "your-nuah-node:9090"  # Проверьте адрес и порт
   timeout = "10s"                   # Увеличьте при медленном соединении
   retry_attempts = 5                # Увеличьте количество попыток
   ```

### 2. Проблемы с торговлей

#### Транзакции отклоняются

**Симптомы:**
- `Transaction failed: insufficient funds`
- `Transaction failed: unauthorized`

**Диагностика:**

1. **Проверка баланса**:
   ```bash
   # Проверка баланса пользователя
   nuahd query bank balances <user-address>

   # Проверка баланса AI-агента
   nuahd query bank balances <ai-agent-address>
   ```

2. **Проверка полномочий**:
   ```bash
   # Проверка authz полномочий
   nuahd query authz grants <user-address> <ai-agent-address>

   # Проверка feegrant
   nuahd query feegrant grants <user-address>
   ```

3. **Проверка лимитов**:
   ```bash
   # Проверка конфигурации лимитов
   grep -A 10 "trading_limits" config.toml
   ```

**Решения:**

1. **Пополнение баланса**:
   ```bash
   # Пополнение баланса пользователя
   nuahd tx bank send <sender> <user-address> 10000factory/test/ndollar
   ```

2. **Обновление полномочий**:
   ```bash
   # Выдача новых полномочий
   nuahd tx authz grant <ai-agent-address> generic \
     --msg-type="/nuah.assets.MsgBuyAsset" \
     --from <user-key>
   ```

3. **Увеличение лимитов**:
   ```toml
   [trading_limits]
   max_daily_volume = "50000factory/test/ndollar"  # Увеличьте лимит
   max_single_trade = "5000factory/test/ndollar"   # Увеличьте лимит
   ```

#### Проблемы с ценами

**Симптомы:**
- `Failed to get price data`
- `Price deviation too high`

**Решения:**

1. **Проверка oracle**:
   ```bash
   # Проверка доступности oracle
   curl http://your-oracle-endpoint/health

   # Проверка ценовых данных
   nuahd query oracle price BTC
   ```

2. **Настройка резервных источников**:
   ```toml
   [data_sources]
   primary_oracle = "nuah-oracle"
   backup_oracles = ["backup-oracle-1", "backup-oracle-2"]
   ```

3. **Увеличение допустимого отклонения**:
   ```toml
   [risk_management]
   max_price_deviation = 0.10  # Увеличьте до 10%
   ```

### 3. Проблемы с мониторингом

#### API не отвечает

**Симптомы:**
- `curl http://localhost:8080/health` не отвечает
- Ошибки 500 в логах

**Диагностика:**

1. **Проверка процесса**:
   ```bash
   # Проверка запущенного процесса
   ps aux | grep monitoring

   # Проверка порта
   netstat -tulpn | grep :8080
   ```

2. **Проверка логов**:
   ```bash
   # Логи мониторинга
   tail -f logs/audit.log
   tail -f logs/alerts.log

   # Системные логи
   journalctl -u ai-trader-monitoring.service -f
   ```

**Решения:**

1. **Перезапуск сервиса**:
   ```bash
   sudo systemctl restart ai-trader-monitoring.service
   ```

2. **Проверка конфигурации**:
   ```toml
   [monitoring]
   port = 8080  # Убедитесь, что порт свободен
   enabled = true
   ```

#### Алерты не отправляются

**Симптомы:**
- Алерты генерируются, но не доходят до получателей
- Ошибки в логах уведомлений

**Диагностика:**

1. **Проверка конфигурации уведомлений**:
   ```toml
   [monitoring.notifiers]
   console_enabled = true
   file_enabled = true
   webhook_enabled = true
   ```

2. **Тестирование webhook**:
   ```bash
   # Тест Slack webhook
   curl -X POST -H 'Content-type: application/json' \
     --data '{"text":"Test alert"}' \
     https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
   ```

**Решения:**

1. **Проверка сетевого подключения**:
   ```bash
   # Проверка доступности внешних сервисов
   ping slack.com
   ping smtp.gmail.com
   ```

2. **Обновление конфигурации**:
   ```toml
   [monitoring.webhook]
   url = "https://hooks.slack.com/services/..."
   timeout = "10s"  # Увеличьте таймаут
   retry_attempts = 5  # Увеличьте попытки
   ```

### 4. Проблемы с производительностью

#### Высокое использование ресурсов

**Симптомы:**
- Высокая загрузка CPU
- Большое потребление памяти
- Медленная работа

**Диагностика:**

```bash
# Мониторинг ресурсов
htop
iotop
netstat -tulpn

# Анализ логов
grep "ERROR" logs/audit.log | wc -l
grep "WARN" logs/audit.log | wc -l
```

**Решения:**

1. **Оптимизация конфигурации**:
   ```toml
   [performance]
   max_concurrent_trades = 3  # Уменьшите количество
   trade_timeout = "60s"      # Увеличьте таймаут
   price_cache_ttl = "120s"   # Увеличьте TTL кэша
   ```

2. **Очистка логов**:
   ```bash
   # Ротация логов
   sudo logrotate -f /etc/logrotate.d/ai-trader

   # Очистка старых логов
   find logs/ -name "*.log.*" -mtime +7 -delete
   ```

3. **Перезапуск сервисов**:
   ```bash
   sudo systemctl restart ai-trader.service
   sudo systemctl restart ai-trader-monitoring.service
   ```

## 🔧 Инструменты диагностики

### Скрипты диагностики

```bash
#!/bin/bash
# diagnostic.sh - Скрипт диагностики AI Trader Bot

echo "=== AI Trader Bot Diagnostic ==="
echo "Timestamp: $(date)"
echo

echo "=== System Status ==="
systemctl status ai-trader.service --no-pager
systemctl status ai-trader-monitoring.service --no-pager
echo

echo "=== Resource Usage ==="
echo "CPU Usage:"
top -bn1 | grep "Cpu(s)"
echo "Memory Usage:"
free -h
echo "Disk Usage:"
df -h
echo

echo "=== Network Status ==="
echo "Listening Ports:"
netstat -tulpn | grep -E ":(8080|9090|26657)"
echo

echo "=== Service Logs (Last 10 lines) ==="
echo "AI Trader Logs:"
journalctl -u ai-trader.service -n 10 --no-pager
echo "Monitoring Logs:"
journalctl -u ai-trader-monitoring.service -n 10 --no-pager
echo

echo "=== Configuration Check ==="
if [ -f "config.toml" ]; then
    echo "Configuration file exists"
    toml validate config.toml 2>/dev/null && echo "Configuration is valid" || echo "Configuration has errors"
else
    echo "Configuration file not found"
fi
echo

echo "=== API Health Check ==="
curl -s http://localhost:8080/health 2>/dev/null && echo "API is responding" || echo "API is not responding"
echo

echo "=== Blockchain Connection ==="
nuahd status --node tcp://localhost:26657 2>/dev/null && echo "Blockchain connection OK" || echo "Blockchain connection failed"
echo

echo "=== Diagnostic Complete ==="
```

### Мониторинг в реальном времени

```bash
#!/bin/bash
# monitor.sh - Мониторинг в реальном времени

while true; do
    clear
    echo "=== AI Trader Bot Monitor ==="
    echo "Timestamp: $(date)"
    echo

    echo "=== Service Status ==="
    systemctl is-active ai-trader.service
    systemctl is-active ai-trader-monitoring.service
    echo

    echo "=== Resource Usage ==="
    echo "CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}')"
    echo "Memory: $(free | grep Mem | awk '{printf "%.1f%%", $3/$2 * 100.0}')"
    echo "Disk: $(df -h / | awk 'NR==2{print $5}')"
    echo

    echo "=== Recent Events ==="
    tail -5 logs/audit.log 2>/dev/null || echo "No audit logs"
    echo

    echo "=== Active Alerts ==="
    curl -s http://localhost:8080/api/v1/alerts?acknowledged=false 2>/dev/null | jq '.total' || echo "API not available"
    echo

    sleep 30
done
```

## 📋 Чек-лист диагностики

### Быстрая диагностика

- [ ] Проверить статус сервисов
- [ ] Проверить логи на ошибки
- [ ] Проверить использование ресурсов
- [ ] Проверить доступность API
- [ ] Проверить подключение к блокчейну
- [ ] Проверить конфигурацию

### Детальная диагностика

- [ ] Проанализировать логи за последние 24 часа
- [ ] Проверить все конфигурационные файлы
- [ ] Протестировать все API эндпоинты
- [ ] Проверить сетевые соединения
- [ ] Проверить права доступа к файлам
- [ ] Протестировать торговые операции

## 🆘 Экстренные процедуры

### Экстренная остановка

```bash
# Остановка всех сервисов
sudo systemctl stop ai-trader.service
sudo systemctl stop ai-trader-monitoring.service

# Отзыв полномочий (если необходимо)
nuahd tx authz revoke <ai-agent-address> /nuah.assets.MsgBuyAsset --from <user-key>
nuahd tx authz revoke <ai-agent-address> /nuah.assets.MsgSellAsset --from <user-key>
nuahd tx feegrant revoke <ai-agent-address> <user-address> --from <user-key>
```

### Восстановление из бэкапа

```bash
# Восстановление конфигурации
cp config.toml.backup config.toml

# Восстановление данных
cp -r data.backup data/

# Перезапуск сервисов
sudo systemctl start ai-trader.service
sudo systemctl start ai-trader-monitoring.service
```

### Переключение на резервную систему

```bash
# Остановка основной системы
sudo systemctl stop ai-trader.service

# Запуск резервной системы
sudo systemctl start ai-trader-backup.service

# Обновление DNS/load balancer
# (зависит от вашей инфраструктуры)
```

## 📞 Получение помощи

### Сбор информации для поддержки

```bash
# Создание отчета о проблеме
./diagnostic.sh > problem_report.txt

# Сбор логов
journalctl -u ai-trader.service --since "1 hour ago" > ai-trader-logs.txt
journalctl -u ai-trader-monitoring.service --since "1 hour ago" > monitoring-logs.txt

# Сбор конфигурации
cp config.toml config-backup.toml
```

### Контакты поддержки

- **Документация**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/nuahchain_osmosis/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/nuahchain_osmosis/discussions)
- **Email**: [support@your-domain.com](mailto:support@your-domain.com)

---

**⚠️ Важно**: Всегда создавайте бэкапы перед внесением изменений и тестируйте решения на тестовой среде перед применением в продакшене.
