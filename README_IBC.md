# IBC Relayer Documentation / Документация IBC релейера

This repository contains comprehensive documentation for setting up and operating an IBC (Inter-Blockchain Communication) relayer using Hermes for the NUAH blockchain.

Этот репозиторий содержит полную документацию по настройке и эксплуатации IBC (Inter-Blockchain Communication) релейера с использованием Hermes для блокчейна NUAH.

## 📚 Documentation Structure / Структура документации

### Main Guides / Основные руководства

- **[English Documentation](docs/IBC_RELAYER_SETUP_EN.md)** - Complete setup guide in English
- **[Русская документация](docs/IBC_RELAYER_SETUP_RU.md)** - Полное руководство по настройке на русском языке
- **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)** - Common issues and solutions / Распространенные проблемы и решения

### Configuration Examples / Примеры конфигурации

- **[Hermes Configuration](examples/hermes_config.toml)** - Complete Hermes configuration file
- **[IBC Commands Reference](examples/ibc_commands.md)** - Common IBC commands and usage examples

### Automation Scripts / Скрипты автоматизации

- **[Setup Script](scripts/setup_ibc_relayer.sh)** - Automated relayer setup script

## 🚀 Quick Start / Быстрый старт

### Prerequisites / Предварительные требования

1. **Rust and Cargo** (for Hermes installation)
2. **Running NUAH node** (local or remote)
3. **Access to target network** (Osmosis testnet or other Cosmos chains)
4. **Sufficient funds** for gas fees on both networks

### Installation / Установка

```bash
# Clone repository / Клонируйте репозиторий
git clone <repository-url>
cd nuahchain_osmosis

# Run automated setup script / Запустите автоматизированный скрипт настройки
chmod +x scripts/setup_ibc_relayer.sh
./scripts/setup_ibc_relayer.sh

# Or follow manual setup in documentation
# Или следуйте ручной настройке в документации
```

## 📖 What's Included / Что включено

### 1. Complete Setup Guides / Полные руководства по настройке

Both English and Russian versions include:
Английская и русская версии включают:

- **Installation instructions** for Hermes relayer
- **Configuration setup** for NUAH and target chains
- **Key management** and security best practices
- **Client, connection, and channel creation**
- **Testing and validation procedures**
- **Monitoring and maintenance guidelines**

### 2. Configuration Templates / Шаблоны конфигурации

- Pre-configured Hermes config for NUAH ↔ Osmosis testnet
- Customizable settings for different network setups
- Security-focused configuration options

### 3. Command Reference / Справочник команд

Comprehensive list of IBC commands including:
Полный список команд IBC, включая:

- Client management / Управление клиентами
- Connection operations / Операции соединений
- Channel management / Управление каналами
- Packet relaying / Релей пакетов
- Transfer operations / Операции трансферов
- Monitoring and debugging / Мониторинг и отладка

### 4. Troubleshooting Guide / Руководство по устранению неполадок

Detailed solutions for:
Подробные решения для:

- Installation issues / Проблемы установки
- Configuration problems / Проблемы конфигурации
- Connection failures / Сбои подключения
- Performance optimization / Оптимизация производительности
- Emergency procedures / Экстренные процедуры

### 5. Automation Tools / Инструменты автоматизации

- **Interactive setup script** with guided configuration
- **Health check utilities** for monitoring
- **Backup and restore procedures**

## 🔧 Supported Networks / Поддерживаемые сети

### Primary Configuration / Основная конфигурация
- **NUAH Chain** (`nuahchain-1`) - Local or remote NUAH blockchain
- **Osmosis Testnet** (`osmo-test-5`) - Public Osmosis testnet

### Extensible to / Расширяемо до
- Any Cosmos SDK-based blockchain
- Custom testnets and mainnets
- Multi-hop IBC connections

## 📋 Features / Функции

### ✅ Automated Setup / Автоматизированная настройка
- One-click relayer installation
- Guided configuration process
- Automatic key generation and management

### ✅ Production Ready / Готово к продакшену
- Security best practices
- Performance optimization
- Monitoring and alerting setup

### ✅ Multi-Language Support / Поддержка нескольких языков
- Complete English documentation
- Полная русская документация
- Bilingual troubleshooting guide

### ✅ Comprehensive Testing / Комплексное тестирование
- End-to-end IBC transfer testing
- Connection validation procedures
- Performance benchmarking tools

## 🛠 Usage Examples / Примеры использования

### Basic IBC Transfer / Базовый IBC трансфер
```bash
# Transfer NUAH tokens to Osmosis
hermes --config ~/.hermes_test/config.toml tx ft-transfer \
  --dst-chain osmo-test-5 \
  --src-chain nuahchain-1 \
  --src-port transfer \
  --src-channel channel-0 \
  --amount 1000 \
  --denom unuah \
  --receiver osmo19rl4cm2hmr8afy4kldpxz3fka4jguq0a5m7df8
```

### Health Check / Проверка состояния
```bash
# Verify relayer configuration and connectivity
hermes --config ~/.hermes_test/config.toml health-check
```

### Start Relaying / Запуск релейинга
```bash
# Start the relayer for all configured chains
hermes --config ~/.hermes_test/config.toml start
```

## 📊 Monitoring / Мониторинг

The documentation includes setup for:
Документация включает настройку для:

- **Real-time logging** and error tracking
- **Balance monitoring** for relayer accounts
- **Performance metrics** and optimization
- **Alert systems** for critical issues

## 🔒 Security Considerations / Соображения безопасности

- **Key management** best practices
- **Network security** configurations
- **Access control** and permissions
- **Backup and recovery** procedures

## 🤝 Contributing / Участие в разработке

To contribute to this documentation:
Чтобы внести вклад в эту документацию:

1. Fork the repository / Сделайте форк репозитория
2. Create a feature branch / Создайте ветку функции
3. Make your changes / Внесите изменения
4. Test the procedures / Протестируйте процедуры
5. Submit a pull request / Отправьте pull request

## 📞 Support / Поддержка

For issues and questions:
По вопросам и проблемам:

- Check the [Troubleshooting Guide](docs/TROUBLESHOOTING.md)
- Review [IBC Commands Reference](examples/ibc_commands.md)
- Open an issue in the repository
- Consult the [Hermes documentation](https://hermes.informal.systems/)

## 📄 License / Лицензия

This documentation is provided under the same license as the NUAH blockchain project.

---

**Note**: This documentation is actively maintained and updated. Please check for the latest version before setting up your IBC relayer.

**Примечание**: Эта документация активно поддерживается и обновляется. Пожалуйста, проверьте последнюю версию перед настройкой вашего IBC релейера.
