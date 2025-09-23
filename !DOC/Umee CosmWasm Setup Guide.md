# Руководство по настройке Umee CosmWasm

## Обзор

Umee CosmWasm - это смарт-контракт, который предоставляет интерфейс для взаимодействия с модулями Umee (Leverage, Oracle, Incentive, Metoken) через CosmWasm. Контракт позволяет выполнять запросы к нативным модулям блокчейна Umee из других CosmWasm контрактов.

## Предварительные требования

### Системные зависимости

1. **Rust** (версия 1.89.0 или выше)
   ```bash
   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
   source ~/.cargo/env
   ```

2. **WebAssembly target**
   ```bash
   rustup target add wasm32-unknown-unknown
   ```

3. **cargo-generate** (для создания новых проектов)
   ```bash
   cargo install cargo-generate
   ```

4. **wasm-opt** (для оптимизации WASM файлов)
   ```bash
   # macOS
   brew install binaryen
   
   # Ubuntu/Debian
   sudo apt install binaryen
   ```

## Установка и сборка

### 1. Клонирование репозитория

```bash
git clone https://github.com/umee-network/umee-cosmwasm.git /path/to/cosmwasm/contracts/umee-cosmwasm
cd /path/to/cosmwasm/contracts/umee-cosmwasm
```

### 2. Исправление совместимости

Контракт использует устаревшие функции CosmWasm. Необходимо заменить:

```bash
# Замена устаревших функций
sed -i 's/from_json/from_binary/g' src/contract.rs
sed -i 's/to_json_binary/to_binary/g' src/contract.rs  
sed -i 's/to_json_vec/to_vec/g' src/contract.rs
```

### 3. Сборка контракта

```bash
# Сборка в режиме release для WebAssembly
cargo build --release --target wasm32-unknown-unknown

# Оптимизация WASM файла
wasm-opt -Oz ../../target/wasm32-unknown-unknown/release/umee_cosmwasm.wasm -o umee_cosmwasm_optimized.wasm
```

## Структура проекта

```
umee-cosmwasm/
├── src/
│   ├── contract.rs      # Основная логика контракта
│   ├── lib.rs          # Библиотечные экспорты
│   ├── msg.rs          # Определения сообщений
│   └── state.rs        # Состояние контракта
├── packages/
│   └── cw-umee-types/  # Типы для взаимодействия с Umee
├── schema/             # JSON схемы для сообщений
├── Cargo.toml         # Конфигурация проекта
└── README.md
```

## Возможности контракта

### Поддерживаемые модули

1. **Leverage** - управление займами и залогами
2. **Oracle** - получение ценовых данных
3. **Incentive** - система поощрений
4. **Metoken** - мета-токены

### Основные функции

#### Запросы (Queries)

- `GetOwner` - получение владельца контракта
- `LeverageParameters` - параметры модуля Leverage
- `RegisteredTokens` - список зарегистрированных токенов
- `AccountBalances` - балансы аккаунта
- `MarketSummary` - сводка по рынку
- `ExchangeRates` - курсы обмена
- `OracleParameters` - параметры Oracle

#### Выполнение (Execute)

- `ChangeOwner` - смена владельца контракта
- Выполнение операций через модуль Leverage

## Размеры файлов

После сборки и оптимизации:
- Исходный WASM: ~654KB
- Оптимизированный WASM: ~517KB (экономия ~21%)

## Развертывание

### Локальное тестирование

```bash
# Запуск тестов
cargo test

# Генерация схем
cargo run --example schema
```

### Развертывание в сети

1. Загрузка контракта:
   ```bash
   osmosisd tx wasm store umee_cosmwasm_optimized.wasm \
     --from <your-key> \
     --gas auto \
     --gas-adjustment 1.3 \
     --chain-id <chain-id>
   ```

2. Инициализация контракта:
   ```bash
   osmosisd tx wasm instantiate <code-id> '{}' \
     --from <your-key> \
     --label "Umee CosmWasm" \
     --gas auto \
     --gas-adjustment 1.3 \
     --chain-id <chain-id>
   ```

## Примеры использования

### Запрос баланса аккаунта

```json
{
  "umee": {
    "leverage": {
      "account_balances": {
        "address": "osmo1..."
      }
    }
  }
}
```

### Получение курсов обмена

```json
{
  "umee": {
    "oracle": {
      "exchange_rates": {
        "denom": "uosmo"
      }
    }
  }
}
```

## Устранение неполадок

### Ошибки компиляции

1. **Unresolved imports**: Убедитесь, что заменили устаревшие функции
2. **Profile warnings**: Нормальные предупреждения для workspace проектов
3. **Snake case warnings**: Косметические предупреждения в типах

### Проблемы с размером

Если WASM файл слишком большой:
1. Используйте `wasm-opt -Oz` для максимальной оптимизации
2. Проверьте неиспользуемые зависимости в `Cargo.toml`
3. Рассмотрите использование `cargo-machete` для очистки

## Дополнительные ресурсы

- [CosmWasm Documentation](https://book.cosmwasm.com/)
- [Umee Protocol](https://umee.cc/)
- [Osmosis Documentation](https://docs.osmosis.zone/)

## Лицензия

Проект распространяется под лицензией, указанной в файле LICENSE репозитория Umee.