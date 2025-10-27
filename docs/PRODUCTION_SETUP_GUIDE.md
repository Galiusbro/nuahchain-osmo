# Production Setup Guide

Руководство по развертыванию AI Trader Bot System в продакшене.

## 🎯 Обзор

Данное руководство описывает процесс развертывания AI Trader Bot System в продакшене с учетом требований безопасности, производительности и надежности.

## 📋 Предварительные требования

### Системные требования

- **CPU**: 4+ ядра
- **RAM**: 8+ GB
- **Диск**: 100+ GB SSD
- **Сеть**: Стабильное соединение с интернетом
- **OS**: Ubuntu 20.04+ / CentOS 8+ / RHEL 8+

### Программное обеспечение

- Go 1.21+
- Docker 20.10+
- Docker Compose 2.0+
- Nginx (для reverse proxy)
- Certbot (для SSL сертификатов)

### Блокчейн требования

- Nuah Chain node (синхронизированная)
- Доступ к gRPC эндпоинту
- Достаточный баланс для комиссий

## 🏗️ Архитектура развертывания

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   AI Trader     │    │   Nuah Node     │
│   (Nginx)       │───▶│   Bot Service   │───▶│   (gRPC)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Monitoring    │    │   Key Storage   │    │   Database      │
│   (Prometheus)  │    │   (Vault/KMS)   │    │   (PostgreSQL)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🔧 Пошаговая установка

### Шаг 1: Подготовка сервера

```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка необходимых пакетов
sudo apt install -y curl wget git build-essential nginx certbot python3-certbot-nginx

# Установка Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

### Шаг 2: Настройка безопасности

```bash
# Создание пользователя для AI trader
sudo useradd -m -s /bin/bash aitrader
sudo usermod -aG docker aitrader

# Настройка SSH ключей
sudo mkdir -p /home/aitrader/.ssh
sudo chmod 700 /home/aitrader/.ssh
sudo chown aitrader:aitrader /home/aitrader/.ssh

# Настройка файрвола
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw allow 8080  # AI Trader API
sudo ufw enable
```

### Шаг 3: Развертывание приложения

```bash
# Переключение на пользователя aitrader
sudo su - aitrader

# Клонирование репозитория
git clone <repository-url> /home/aitrader/ai-trader-bot
cd /home/aitrader/ai-trader-bot

# Сборка приложения
go mod download
go build -o bin/ai-trader ./services/ai_trader/main.go
go build -o bin/monitoring ./services/ai_trader/monitoring/example/main.go
```

### Шаг 4: Конфигурация

```bash
# Создание директорий
mkdir -p /home/aitrader/ai-trader-bot/{config,logs,data}

# Создание production конфигурации
cat > /home/aitrader/ai-trader-bot/config/production.toml << 'EOF'
[bot]
name = "production-ai-trader"
version = "1.0.0"
environment = "production"

[logging]
level = "info"
format = "json"
audit_log_path = "/home/aitrader/ai-trader-bot/logs/audit.log"
alert_log_path = "/home/aitrader/ai-trader-bot/logs/alerts.log"
max_file_size = 104857600  # 100MB
max_files = 10

[trading_limits]
max_daily_volume = "50000factory/test/ndollar"
max_trades_per_day = 500
max_trades_per_hour = 50
max_single_trade = "5000factory/test/ndollar"

[risk_management]
max_price_deviation = 0.03
stop_loss_percentage = 0.05
take_profit_percentage = 0.15
cooldown_period = "2m"
max_consecutive_losses = 3
max_daily_loss = "10000factory/test/ndollar"

[oracle]
endpoint = "your-nuah-node:9090"
timeout = "5s"
retry_attempts = 5
retry_delay = "1s"

[monitoring]
enabled = true
port = 8080
metrics_enabled = true
health_check_interval = "30s"

[security]
key_storage_type = "vault"  # vault, kms, file
vault_endpoint = "https://vault.your-domain.com"
vault_token_path = "/home/aitrader/.vault-token"
EOF
```

### Шаг 5: Настройка systemd сервисов

```bash
# Создание systemd сервиса для AI Trader
sudo tee /etc/systemd/system/ai-trader.service > /dev/null << 'EOF'
[Unit]
Description=AI Trader Bot Service
After=network.target

[Service]
Type=simple
User=aitrader
Group=aitrader
WorkingDirectory=/home/aitrader/ai-trader-bot
ExecStart=/home/aitrader/ai-trader-bot/bin/ai-trader --config /home/aitrader/ai-trader-bot/config/production.toml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/home/aitrader/ai-trader-bot/logs
ReadWritePaths=/home/aitrader/ai-trader-bot/data

[Install]
WantedBy=multi-user.target
EOF

# Создание systemd сервиса для мониторинга
sudo tee /etc/systemd/system/ai-trader-monitoring.service > /dev/null << 'EOF'
[Unit]
Description=AI Trader Monitoring Service
After=network.target ai-trader.service

[Service]
Type=simple
User=aitrader
Group=aitrader
WorkingDirectory=/home/aitrader/ai-trader-bot
ExecStart=/home/aitrader/ai-trader-bot/bin/monitoring
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/home/aitrader/ai-trader-bot/logs

[Install]
WantedBy=multi-user.target
EOF

# Перезагрузка systemd и запуск сервисов
sudo systemctl daemon-reload
sudo systemctl enable ai-trader.service
sudo systemctl enable ai-trader-monitoring.service
```

### Шаг 6: Настройка Nginx

```bash
# Создание конфигурации Nginx
sudo tee /etc/nginx/sites-available/ai-trader > /dev/null << 'EOF'
server {
    listen 80;
    server_name your-domain.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL configuration
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;

    # AI Trader API
    location /api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # Health check
    location /health {
        proxy_pass http://localhost:8080/health;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Static files (if any)
    location /static/ {
        alias /home/aitrader/ai-trader-bot/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
EOF

# Активация сайта
sudo ln -s /etc/nginx/sites-available/ai-trader /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Шаг 7: SSL сертификаты

```bash
# Получение SSL сертификата
sudo certbot --nginx -d your-domain.com

# Настройка автоматического обновления
sudo crontab -e
# Добавить строку:
# 0 12 * * * /usr/bin/certbot renew --quiet
```

### Шаг 8: Настройка мониторинга

```bash
# Установка Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xzf prometheus-2.45.0.linux-amd64.tar.gz
sudo mv prometheus-2.45.0.linux-amd64 /opt/prometheus

# Создание конфигурации Prometheus
sudo tee /opt/prometheus/prometheus.yml > /dev/null << 'EOF'
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'ai-trader'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/api/v1/metrics'
    scrape_interval: 30s
EOF

# Создание systemd сервиса для Prometheus
sudo tee /etc/systemd/system/prometheus.service > /dev/null << 'EOF'
[Unit]
Description=Prometheus
After=network.target

[Service]
Type=simple
User=prometheus
Group=prometheus
ExecStart=/opt/prometheus/prometheus --config.file=/opt/prometheus/prometheus.yml --storage.tsdb.path=/opt/prometheus/data
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo useradd --no-create-home --shell /bin/false prometheus
sudo chown -R prometheus:prometheus /opt/prometheus
sudo systemctl enable prometheus
```

## 🔐 Безопасность

### Управление ключами

#### Вариант 1: HashiCorp Vault

```bash
# Установка Vault
wget https://releases.hashicorp.com/vault/1.15.2/vault_1.15.2_linux_amd64.zip
unzip vault_1.15.2_linux_amd64.zip
sudo mv vault /usr/local/bin/

# Инициализация Vault
vault operator init

# Сохранение ключей в Vault
vault kv put secret/ai-trader/user-key value="<user-private-key>"
vault kv put secret/ai-trader/ai-agent-key value="<ai-agent-private-key>"
```

#### Вариант 2: AWS KMS

```bash
# Установка AWS CLI
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Настройка AWS credentials
aws configure

# Создание KMS ключа
aws kms create-key --description "AI Trader Bot Keys"
```

### Настройка файрвола

```bash
# Дополнительные правила файрвола
sudo ufw deny 9090  # Блокировка Prometheus извне
sudo ufw deny 8080  # Блокировка AI Trader API извне
sudo ufw allow from 10.0.0.0/8 to any port 8080  # Разрешить только из внутренней сети
```

## 📊 Мониторинг и логирование

### Настройка логирования

```bash
# Настройка logrotate
sudo tee /etc/logrotate.d/ai-trader > /dev/null << 'EOF'
/home/aitrader/ai-trader-bot/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 aitrader aitrader
    postrotate
        systemctl reload ai-trader.service
        systemctl reload ai-trader-monitoring.service
    endscript
}
EOF
```

### Настройка алертов

```bash
# Установка Alertmanager
wget https://github.com/prometheus/alertmanager/releases/download/v0.25.0/alertmanager-0.25.0.linux-amd64.tar.gz
tar xzf alertmanager-0.25.0.linux-amd64.tar.gz
sudo mv alertmanager-0.25.0.linux-amd64 /opt/alertmanager

# Создание конфигурации алертов
sudo tee /opt/alertmanager/alertmanager.yml > /dev/null << 'EOF'
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@your-domain.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  email_configs:
  - to: 'admin@your-domain.com'
    subject: 'AI Trader Alert: {{ .GroupLabels.alertname }}'
    body: |
      {{ range .Alerts }}
      Alert: {{ .Annotations.summary }}
      Description: {{ .Annotations.description }}
      {{ end }}
EOF
```

## 🚀 Запуск и проверка

### Запуск сервисов

```bash
# Запуск всех сервисов
sudo systemctl start ai-trader.service
sudo systemctl start ai-trader-monitoring.service
sudo systemctl start prometheus.service
sudo systemctl start alertmanager.service

# Проверка статуса
sudo systemctl status ai-trader.service
sudo systemctl status ai-trader-monitoring.service
```

### Проверка работоспособности

```bash
# Проверка API
curl -k https://your-domain.com/health

# Проверка метрик
curl -k https://your-domain.com/api/v1/metrics

# Проверка логов
sudo journalctl -u ai-trader.service -f
tail -f /home/aitrader/ai-trader-bot/logs/audit.log
```

## 🔄 Обновление и обслуживание

### Обновление приложения

```bash
# Создание бэкапа
sudo systemctl stop ai-trader.service
cp -r /home/aitrader/ai-trader-bot /home/aitrader/ai-trader-bot.backup.$(date +%Y%m%d)

# Обновление кода
cd /home/aitrader/ai-trader-bot
git pull origin main
go build -o bin/ai-trader ./services/ai_trader/main.go

# Перезапуск сервиса
sudo systemctl start ai-trader.service
```

### Мониторинг производительности

```bash
# Мониторинг ресурсов
htop
iotop
netstat -tulpn | grep :8080

# Проверка дискового пространства
df -h
du -sh /home/aitrader/ai-trader-bot/logs/*
```

## 🆘 Устранение неполадок

### Частые проблемы

1. **Сервис не запускается**:
   ```bash
   sudo journalctl -u ai-trader.service -n 50
   ```

2. **Проблемы с подключением к блокчейну**:
   ```bash
   # Проверка доступности ноды
   telnet your-nuah-node 9090
   ```

3. **Высокое использование памяти**:
   ```bash
   # Проверка утечек памяти
   sudo systemctl restart ai-trader.service
   ```

### Логи и диагностика

```bash
# Просмотр логов в реальном времени
sudo journalctl -u ai-trader.service -f

# Анализ логов аудита
grep "ERROR" /home/aitrader/ai-trader-bot/logs/audit.log

# Проверка конфигурации
/home/aitrader/ai-trader-bot/bin/ai-trader --config /home/aitrader/ai-trader-bot/config/production.toml --validate
```

## 📋 Чек-лист развертывания

- [ ] Сервер подготовлен и настроен
- [ ] Безопасность настроена (файрвол, SSL, пользователи)
- [ ] Приложение собрано и установлено
- [ ] Конфигурация создана и проверена
- [ ] Systemd сервисы настроены и запущены
- [ ] Nginx настроен и работает
- [ ] SSL сертификаты получены
- [ ] Мониторинг настроен
- [ ] Ключи безопасно сохранены
- [ ] Тесты пройдены
- [ ] Документация обновлена
- [ ] Команда обучена

---

**⚠️ Важно**: Данное руководство является базовым. Для продакшена требуется дополнительная настройка безопасности, мониторинга и резервного копирования в соответствии с требованиями вашей организации.
