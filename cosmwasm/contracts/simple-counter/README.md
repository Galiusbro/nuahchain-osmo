# Simple Counter Contract

Простой CosmWasm смарт-контракт счетчика для демонстрации основных возможностей.

## Описание

Этот контракт реализует простой счетчик с возможностью:
- Инициализации с начальным значением
- Увеличения счетчика на 1
- Сброса счетчика (только владельцем)
- Запроса текущего значения

## Сообщения

### InstantiateMsg
```json
{
  "count": 0
}
```

### ExecuteMsg
```json
// Увеличить счетчик на 1
{
  "increment": {}
}

// Сбросить счетчик (только владелец)
{
  "reset": {
    "count": 42
  }
}
```

### QueryMsg
```json
// Получить текущее значение счетчика
{
  "get_count": {}
}
```

## Компиляция

```bash
# Установить WASM target
rustup target add wasm32-unknown-unknown

# Скомпилировать контракт
cargo build --release --target wasm32-unknown-unknown

# Запустить тесты
cargo test

# Генерировать схему
cargo run --example schema
```

## Файлы

- `src/contract.rs` - Основная логика контракта
- `src/msg.rs` - Определения сообщений
- `src/state.rs` - Структуры состояния
- `src/error.rs` - Типы ошибок
- `tests/integration.rs` - Интеграционные тесты
- `examples/schema.rs` - Генератор JSON схемы

## Результат компиляции

WASM файл: `../../target/wasm32-unknown-unknown/release/simple_counter.wasm`
Размер: ~189KB

## Тестирование

Контракт включает:
- Unit тесты в `src/contract.rs`
- Интеграционные тесты в `tests/integration.rs`
- Тестирование авторизации и ошибок