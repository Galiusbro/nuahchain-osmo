# Monitoring Guide

Руководство по мониторингу и аудиту AI Trader Bot System.

## 🎯 Обзор

Система мониторинга AI Trader Bot обеспечивает полную прозрачность торговых операций, отслеживание производительности и контроль соблюдения политик безопасности.

## 📊 Компоненты мониторинга

### 1. Логирование событий
- **Audit Logs**: Полный журнал всех торговых операций
- **Alert Logs**: Записи о нарушениях и предупреждениях
- **System Logs**: Системные события и ошибки

### 2. Метрики производительности
- **Trading Metrics**: Объем торгов, количество сделок, P&L
- **Risk Metrics**: Нарушения лимитов, срабатывания стоп-лоссов
- **System Metrics**: Время отклика, доступность сервисов

### 3. Алерты и уведомления
- **Policy Violations**: Нарушения торговых политик
- **Risk Alerts**: Превышение лимитов риска
- **System Alerts**: Технические проблемы

## 🔧 Настройка мониторинга

### Базовая конфигурация

```toml
[monitoring]
enabled = true
port = 8080
metrics_enabled = true
health_check_interval = "30s"

[logging]
level = "info"
format = "json"
audit_log_path = "logs/audit.log"
alert_log_path = "logs/alerts.log"
buffer_size = 1000
flush_interval = "30s"
```

### Расширенная конфигурация

```toml
[monitoring]
enabled = true
port = 8080
metrics_enabled = true
health_check_interval = "30s"
alert_rules_enabled = true

[monitoring.notifiers]
console_enabled = true
file_enabled = true
webhook_enabled = true
email_enabled = true

[monitoring.webhook]
url = "https://hooks.slack.com/services/..."
headers = { "Content-Type" = "application/json" }
timeout = "5s"

[monitoring.email]
smtp_host = "smtp.gmail.com"
smtp_port = 587
from = "alerts@your-domain.com"
to = ["admin@your-domain.com", "trader@your-domain.com"]
```

## 📈 API Эндпоинты

### Health Check

```bash
# Проверка состояния системы
curl http://localhost:8080/health

# Ответ:
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "uptime": "2h30m15s"
}
```

### Events API

```bash
# Получение всех событий
curl "http://localhost:8080/api/v1/events"

# Фильтрация по трейдеру
curl "http://localhost:8080/api/v1/events?trader=cosmos1..."

# Фильтрация по типу события
curl "http://localhost:8080/api/v1/events?event_type=trade_executed"

# Фильтрация по уровню риска
curl "http://localhost:8080/api/v1/events?risk_level=high"

# Пагинация
curl "http://localhost:8080/api/v1/events?offset=0&limit=100"
```

### Alerts API

```bash
# Получение всех алертов
curl "http://localhost:8080/api/v1/alerts"

# Фильтрация по серьезности
curl "http://localhost:8080/api/v1/alerts?severity=critical"

# Фильтрация по типу
curl "http://localhost:8080/api/v1/alerts?alert_type=policy_violation"

# Подтверждение алерта
curl -X POST "http://localhost:8080/api/v1/alerts/alert-id/acknowledge"
```

### History API

```bash
# История трейдера
curl "http://localhost:8080/api/v1/history/cosmos1..."

# История по символу
curl "http://localhost:8080/api/v1/history/cosmos1.../BTC"
```

### Metrics API

```bash
# Системные метрики
curl "http://localhost:8080/api/v1/metrics"

# Метрики трейдера
curl "http://localhost:8080/api/v1/metrics/trader/cosmos1..."
```

## 📋 Типы событий

### Trade Events

```json
{
  "id": "evt_1234567890",
  "timestamp": "2024-01-01T12:00:00Z",
  "event_type": "trade_executed",
  "trader": "cosmos1...",
  "symbol": "BTC",
  "action": "buy",
  "amount": "1000factory/test/ndollar",
  "price": "50000.00",
  "tx_hash": "0x1234567890abcdef",
  "success": true,
  "risk_level": "low"
}
```

### Policy Violation Events

```json
{
  "id": "evt_1234567891",
  "timestamp": "2024-01-01T12:00:00Z",
  "event_type": "policy_violation",
  "trader": "cosmos1...",
  "symbol": "BTC",
  "reason": "Volume limit exceeded",
  "violations": ["daily_volume_limits"],
  "risk_level": "high"
}
```

### Emergency Stop Events

```json
{
  "id": "evt_1234567892",
  "timestamp": "2024-01-01T12:00:00Z",
  "event_type": "emergency_stop",
  "trader": "cosmos1...",
  "reason": "Consecutive losses limit exceeded",
  "metadata": {
    "consecutive_losses": 5,
    "daily_loss": "1000000"
  },
  "risk_level": "critical"
}
```

## 🚨 Типы алертов

### Policy Violation Alerts

```json
{
  "id": "alert_1234567890",
  "timestamp": "2024-01-01T12:00:00Z",
  "alert_type": "policy_violation",
  "severity": "warning",
  "title": "Policy Violation",
  "message": "Volume limit exceeded for trader cosmos1...",
  "trader": "cosmos1...",
  "symbol": "BTC",
  "metadata": {
    "violations": ["daily_volume_limits"],
    "current_volume": "1500000",
    "limit": "1000000"
  }
}
```

### Emergency Stop Alerts

```json
{
  "id": "alert_1234567891",
  "timestamp": "2024-01-01T12:00:00Z",
  "alert_type": "emergency_stop",
  "severity": "critical",
  "title": "Emergency Stop Triggered",
  "message": "Emergency stop activated for trader cosmos1...",
  "trader": "cosmos1...",
  "metadata": {
    "reason": "consecutive_losses",
    "consecutive_losses": 5,
    "max_allowed": 3
  }
}
```

### System Error Alerts

```json
{
  "id": "alert_1234567892",
  "timestamp": "2024-01-01T12:00:00Z",
  "alert_type": "system_error",
  "severity": "error",
  "title": "Oracle Connection Failed",
  "message": "Failed to connect to oracle endpoint",
  "metadata": {
    "endpoint": "localhost:9090",
    "error": "connection timeout",
    "retry_attempts": 3
  }
}
```

## 📊 Метрики и KPI

### Trading Metrics

```json
{
  "total_trades": 1250,
  "successful_trades": 1180,
  "failed_trades": 70,
  "total_volume": "50000000factory/test/ndollar",
  "total_profit": "2500000factory/test/ndollar",
  "total_loss": "500000factory/test/ndollar",
  "success_rate": 0.944,
  "average_trade_size": "40000factory/test/ndollar"
}
```

### Risk Metrics

```json
{
  "policy_violations": 15,
  "emergency_stops": 2,
  "stop_loss_triggers": 8,
  "take_profit_triggers": 25,
  "max_drawdown": "1000000factory/test/ndollar",
  "volatility": 0.15,
  "sharpe_ratio": 1.8
}
```

### System Metrics

```json
{
  "uptime": "99.9%",
  "average_response_time": "150ms",
  "oracle_availability": "99.5%",
  "active_traders": 5,
  "active_alerts": 3,
  "memory_usage": "512MB",
  "cpu_usage": "25%"
}
```

## 🔔 Настройка алертов

### Правила алертов

```toml
[monitoring.alert_rules]
# Правило превышения дневного объема
[monitoring.alert_rules.high_volume]
enabled = true
condition = "volume_threshold"
threshold = 1000000
time_window = "1h"
severity = "warning"

# Правило превышения количества сделок
[monitoring.alert_rules.high_frequency]
enabled = true
condition = "trade_count"
threshold = 50
time_window = "1h"
severity = "warning"

# Правило подряд идущих убытков
[monitoring.alert_rules.consecutive_losses]
enabled = true
condition = "consecutive_losses"
threshold = 5
time_window = "24h"
severity = "error"

# Правило недоступности oracle
[monitoring.alert_rules.oracle_down]
enabled = true
condition = "oracle_unavailable"
threshold = 1
time_window = "5m"
severity = "critical"
```

### Уведомления

#### Console Notifications

```toml
[monitoring.notifiers.console]
enabled = true
format = "detailed"  # detailed, simple
```

#### File Notifications

```toml
[monitoring.notifiers.file]
enabled = true
path = "logs/alerts.log"
format = "json"
```

#### Webhook Notifications

```toml
[monitoring.notifiers.webhook]
enabled = true
url = "https://hooks.slack.com/services/..."
headers = { "Content-Type" = "application/json" }
timeout = "5s"
retry_attempts = 3
```

#### Email Notifications

```toml
[monitoring.notifiers.email]
enabled = true
smtp_host = "smtp.gmail.com"
smtp_port = 587
username = "alerts@your-domain.com"
password = "your-app-password"
from = "alerts@your-domain.com"
to = ["admin@your-domain.com", "trader@your-domain.com"]
subject_template = "AI Trader Alert: {{ .AlertType }}"
```

## 📈 Дашборды и визуализация

### Prometheus Integration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'ai-trader'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/api/v1/metrics'
    scrape_interval: 30s
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "AI Trader Bot Dashboard",
    "panels": [
      {
        "title": "Trading Volume",
        "type": "graph",
        "targets": [
          {
            "expr": "ai_trader_total_volume",
            "legendFormat": "Total Volume"
          }
        ]
      },
      {
        "title": "Success Rate",
        "type": "singlestat",
        "targets": [
          {
            "expr": "ai_trader_success_rate",
            "legendFormat": "Success Rate"
          }
        ]
      },
      {
        "title": "Active Alerts",
        "type": "table",
        "targets": [
          {
            "expr": "ai_trader_active_alerts",
            "legendFormat": "Alerts"
          }
        ]
      }
    ]
  }
}
```

## 🔍 Анализ и отчетность

### Ежедневные отчеты

```bash
# Генерация ежедневного отчета
curl "http://localhost:8080/api/v1/reports/daily?date=2024-01-01" > daily_report.json

# Отчет по трейдеру
curl "http://localhost:8080/api/v1/reports/trader/cosmos1...?period=7d" > trader_report.json
```

### Анализ производительности

```bash
# Анализ торговой активности
curl "http://localhost:8080/api/v1/analytics/trading?period=30d"

# Анализ рисков
curl "http://localhost:8080/api/v1/analytics/risk?period=7d"

# Анализ системы
curl "http://localhost:8080/api/v1/analytics/system?period=24h"
```

## 🛠️ Устранение неполадок

### Диагностика мониторинга

```bash
# Проверка состояния мониторинга
curl http://localhost:8080/health

# Проверка логов
tail -f logs/audit.log
tail -f logs/alerts.log

# Проверка метрик
curl http://localhost:8080/api/v1/metrics

# Проверка алертов
curl http://localhost:8080/api/v1/alerts
```

### Частые проблемы

1. **Мониторинг не запускается**:
   - Проверьте конфигурацию
   - Убедитесь, что порт 8080 свободен
   - Проверьте права доступа к файлам логов

2. **Алерты не отправляются**:
   - Проверьте настройки уведомлений
   - Убедитесь в доступности внешних сервисов
   - Проверьте логи на ошибки

3. **Метрики не обновляются**:
   - Проверьте подключение к Prometheus
   - Убедитесь в правильности эндпоинта
   - Проверьте интервал сбора метрик

### Команды диагностики

```bash
# Проверка портов
netstat -tulpn | grep :8080

# Проверка процессов
ps aux | grep monitoring

# Проверка дискового пространства
df -h logs/

# Проверка прав доступа
ls -la logs/
```

## 📚 Интеграция с внешними системами

### Slack Integration

```bash
# Создание Slack webhook
curl -X POST -H 'Content-type: application/json' \
  --data '{"text":"AI Trader Alert: Policy violation detected"}' \
  https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
```

### PagerDuty Integration

```bash
# Создание PagerDuty alert
curl -X POST https://events.pagerduty.com/v2/enqueue \
  -H 'Content-Type: application/json' \
  -d '{
    "routing_key": "your-routing-key",
    "event_action": "trigger",
    "payload": {
      "summary": "AI Trader Emergency Stop",
      "severity": "critical",
      "source": "ai-trader-bot"
    }
  }'
```

### Custom Webhook

```bash
# Отправка данных в внешнюю систему
curl -X POST https://your-monitoring-system.com/api/alerts \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer your-token' \
  -d '{
    "alert_id": "alert_1234567890",
    "severity": "warning",
    "message": "Policy violation detected"
  }'
```

---

**⚠️ Важно**: Регулярно проверяйте работоспособность системы мониторинга и обновляйте правила алертов в соответствии с изменяющимися требованиями к торговле.

