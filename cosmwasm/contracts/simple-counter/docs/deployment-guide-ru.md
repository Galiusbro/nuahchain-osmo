# Руководство по развертыванию смарт-контракта Simple Counter

Это руководство описывает пошаговый процесс компиляции, развертывания и тестирования смарт-контракта Simple Counter на блокчейне Nuah Chain.

## Предварительные требования

- Запущенная нода Nuah Chain
- Установленный Rust с поддержкой WebAssembly
- Инструменты для оптимизации WASM (wasm-opt)
- Аккаунт с достаточным балансом для оплаты комиссий

## Шаг 1: Проверка статуса ноды

Убедитесь, что нода Nuah Chain запущена и работает:

```bash
./build/nuahd status
```

## Шаг 2: Компиляция контракта

### 2.1 Установка старой версии Rust (для совместимости)

```bash
rustup toolchain install 1.65.0
rustup target add wasm32-unknown-unknown --toolchain 1.65.0
```

### 2.2 Компиляция контракта

Перейдите в директорию контракта:

```bash
cd cosmwasm/contracts/simple-counter
```

Скомпилируйте контракт с использованием старой версии Rust:

```bash
RUSTFLAGS='-C link-arg=-s' cargo +1.65.0 build --release --target wasm32-unknown-unknown --locked
```

### 2.3 Копирование скомпилированного файла

```bash
cp ../../target/wasm32-unknown-unknown/release/simple_counter.wasm simple_counter_old_rust.wasm
```

## Шаг 3: Подготовка аккаунта

### 3.1 Просмотр доступных ключей

```bash
./build/nuahd keys list
```

### 3.2 Проверка баланса аккаунта

```bash
./build/nuahd query bank balances <адрес_аккаунта>
```

### 3.3 Пополнение баланса (если необходимо)

Найдите аккаунт с балансом в genesis файле:

```bash
cat ~/.nuahd/config/genesis.json | jq '.app_state.bank.balances'
```

Переведите средства на ваш аккаунт:

```bash
./build/nuahd tx bank send <адрес_отправителя> <адрес_получателя> 1000000unuah \
  --chain-id nuahchain-1 --fees 5000unuah --yes --keyring-backend test
```

## Шаг 4: Развертывание контракта

### 4.1 Загрузка кода контракта

```bash
./build/nuahd tx wasm store cosmwasm/contracts/simple-counter/simple_counter_old_rust.wasm \
  --from alice --chain-id nuahchain-1 --gas auto --gas-adjustment 1.3 \
  --fees 50000unuah --yes
```

### 4.2 Проверка загруженного кода

Подождите несколько секунд и проверьте список загруженных контрактов:

```bash
./build/nuahd query wasm list-code
```

Вы должны увидеть ваш контракт с `code_id: "1"`.

## Шаг 5: Создание экземпляра контракта

### 5.1 Инициализация контракта

```bash
./build/nuahd tx wasm instantiate 1 '{"count": 10}' \
  --from alice --chain-id nuahchain-1 --label "simple-counter" \
  --admin <адрес_администратора> --gas auto --gas-adjustment 1.3 \
  --fees 30000unuah --yes
```

### 5.2 Получение адреса контракта

```bash
./build/nuahd query wasm list-contract-by-code 1
```

Сохраните полученный адрес контракта для дальнейшего использования.

## Шаг 6: Тестирование контракта

### 6.1 Запрос текущего значения счетчика

```bash
./build/nuahd query wasm contract-state smart <адрес_контракта> '{"get_count":{}}'
```

Ожидаемый результат:
```yaml
data:
  count: 10
```

### 6.2 Увеличение счетчика

```bash
./build/nuahd tx wasm execute <адрес_контракта> '{"increment":{}}' \
  --from alice --chain-id nuahchain-1 --gas auto --gas-adjustment 1.3 \
  --fees 20000unuah --yes
```

Подождите несколько секунд и проверьте новое значение:

```bash
./build/nuahd query wasm contract-state smart <адрес_контракта> '{"get_count":{}}'
```

Ожидаемый результат:
```yaml
data:
  count: 11
```

### 6.3 Сброс счетчика

```bash
./build/nuahd tx wasm execute <адрес_контракта> '{"reset":{"count":5}}' \
  --from alice --chain-id nuahchain-1 --gas auto --gas-adjustment 1.3 \
  --fees 20000unuah --yes
```

Проверьте результат:

```bash
./build/nuahd query wasm contract-state smart <адрес_контракта> '{"get_count":{}}'
```

Ожидаемый результат:
```yaml
data:
  count: 5
```

## Возможные проблемы и решения

### Проблема: "bulk memory operations not enabled"

**Решение**: Используйте старую версию Rust (1.65.0) для компиляции, как показано в шаге 2.1.

### Проблема: "key not found" или "account not found"

**Решение**: Убедитесь, что у аккаунта есть достаточный баланс. Используйте шаг 3 для пополнения баланса.

### Проблема: "you must set an admin or explicitly pass --no-admin"

**Решение**: Добавьте параметр `--admin <адрес>` при инициализации контракта или используйте `--no-admin` для создания неизменяемого контракта.

### Проблема: "insufficient fee"

**Решение**: Увеличьте размер комиссии, добавив параметр `--fees <сумма>unuah` к команде.

## Заключение

После выполнения всех шагов у вас будет полностью функционирующий смарт-контракт Simple Counter, развернутый на блокчейне Nuah Chain. Контракт поддерживает три основные операции:

1. `get_count` - получение текущего значения счетчика
2. `increment` - увеличение счетчика на 1
3. `reset` - установка счетчика в определенное значение

Вы можете использовать полученный адрес контракта для дальнейшего взаимодействия с ним через CLI или интеграции в другие приложения.
