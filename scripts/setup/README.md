# Setup Scripts

Эта директория содержит скрипты для настройки и тестирования узла nuahchain.

## Основные скрипты

### `init_fresh_node.sh`
Полная инициализация узла с нуля:
- Удаляет все существующие данные и ключи
- Создает новые ключи (validator, alice, bob)
- Настраивает генезис с начальными балансами
- Подготавливает узел к запуску

**Использование:**
```bash
./scripts/setup/init_fresh_node.sh
```

### `test_fee_transaction.sh`
Тестирование транзакций с комиссией в ndollar:
- Проверяет балансы аккаунтов
- Выполняет dry-run транзакции
- Отправляет реальную транзакцию с комиссией в ndollar
- Проверяет результат

**Использование:**
```bash
# Сначала запустите узел
./build/nuahd start

# В другом терминале запустите тест
./scripts/setup/test_fee_transaction.sh
```

## Существующие скрипты

### Настройка различных компонентов
- `setup_ndollar.sh` - Настройка токена ndollar
- `setup_fee_abstraction.sh` - Настройка абстракции комиссий
- `setup_superfluid.sh` - Настройка superfluid
- `setup_production_tokenomics.sh` - Настройка продакшн токеномики
- `setup_proper_tokenomics.sh` - Настройка правильной токеномики

### Настройка сети
- `setup-testnet.sh` - Настройка тестнета
- `setup-droplet.sh` - Настройка на droplet
- `multinode-local-testnet.sh` - Локальный мультинодовый тестнет

### IBC и реле
- `setup_ibc_relayer.sh` - Настройка IBC релея
- `setup_ibc_relayer_test.sh` - Тестирование IBC релея

## Быстрый старт

1. Инициализируйте новый узел:
   ```bash
   ./scripts/setup/init_fresh_node.sh
   ```

2. Запустите узел:
   ```bash
   ./build/nuahd start
   ```

3. В другом терминале протестируйте транзакции:
   ```bash
   ./scripts/setup/test_fee_transaction.sh
   ```

## Примечания

- Все скрипты используют keyring-backend "test" для удобства разработки
- Chain ID по умолчанию: "nuahchain"
- Начальные балансы включают как unuah, так и ndollar токены
- Скрипты автоматически создают необходимые аккаунты и настройки
