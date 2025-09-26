# Инструкция по развертыванию ноды блокчейна NUAH

## 📋 Предварительные требования

### Системные требования
- Ubuntu 20.04+ или аналогичная Linux система
- Минимум 4GB RAM
- Минимум 100GB свободного места на диске
- Стабильное интернет-соединение
- Открытые порты: 26657 (RPC), 1317 (REST API), 9090 (gRPC)

### Необходимые инструменты
- Go 1.21+
- Git
- Make
- SSH доступ к серверу

## 🚀 Пошаговая инструкция развертывания

### 1. Подготовка сервера

```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка необходимых пакетов
sudo apt install -y build-essential git curl wget jq

# Установка Go (если не установлен)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. Перенос кода на сервер

```bash
# Создание архива проекта (выполнить локально)
tar -czf nuahchain_osmosis.tar.gz nuahchain_osmosis/

# Копирование на сервер
scp -i /path/to/your/key nuahchain_osmosis.tar.gz user@your-server-ip:~/

# На сервере: распаковка архива
ssh -i /path/to/your/key user@your-server-ip
tar -xzf nuahchain_osmosis.tar.gz
cd nuahchain_osmosis
```

### 3. Сборка бинарника

```bash
# Сборка проекта
make build

# Проверка успешной сборки
ls -la build/
./build/nuahd version
```

### 4. Инициализация ноды

```bash
# Инициализация ноды с указанием chain-id и moniker
./build/nuahd init test-node --chain-id nuahchain-1

# Проверка создания конфигурационных файлов
ls -la ~/.nuahd/config/
```

### 5. Создание ключа валидатора

```bash
# Создание ключа валидатора
./build/nuahd keys add validator --keyring-backend test

# ВАЖНО: Сохраните мнемоническую фразу в безопасном месте!
# Пример вывода:
# - name: validator
# - type: local
# - address: nuah1...
# - pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"..."}'
# - mnemonic: "word1 word2 word3 ... word24"
```

### 6. Настройка токена

```bash
# Изменение деноминации с 'stake' на 'unuah' в genesis.json
sed -i 's/"stake"/"unuah"/g' ~/.nuahd/config/genesis.json

# Проверка изменений
grep -n "unuah" ~/.nuahd/config/genesis.json
```

### 7. Создание генезис аккаунта

```bash
# Добавление генезис аккаунта с 10 миллионами токенов NUAH
./build/nuahd add-genesis-account validator 10000000000000unuah --keyring-backend test

# Проверка добавления аккаунта
./build/nuahd query bank balances $(./build/nuahd keys show validator --keyring-backend test -a) --genesis
```

### 8. Создание генезис транзакции

```bash
# Создание генезис транзакции с 5 миллионами токенов для стейкинга
./build/nuahd gentx validator 5000000000000unuah --chain-id nuahchain-1 --keyring-backend test

# Сбор всех генезис транзакций
./build/nuahd collect-gentxs

# Проверка валидности генезиса
./build/nuahd validate-genesis
```

### 9. Настройка файрвола

```bash
# Включение UFW (если не включен)
sudo ufw enable

# Открытие необходимых портов
sudo ufw allow 26657/tcp  # RPC порт
sudo ufw allow 1317/tcp   # REST API порт
sudo ufw allow 9090/tcp   # gRPC порт
sudo ufw allow 22/tcp     # SSH порт

# Проверка статуса файрвола
sudo ufw status
```

### 10. Запуск ноды

```bash
# Запуск ноды в фоновом режиме с внешним доступом RPC
nohup ./build/nuahd start --rpc.laddr=tcp://0.0.0.0:26657 > nuahd.log 2>&1 &

# Проверка запуска
ps aux | grep nuahd
```

## 🔍 Проверка работы ноды

### Проверка логов
```bash
# Просмотр последних логов
tail -f nuahd.log

# Просмотр последних 20 строк
tail -20 nuahd.log
```

### Проверка статуса через RPC
```bash
# Проверка статуса ноды
curl -s http://localhost:26657/status | jq

# Проверка высоты блока
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'
```

### Проверка баланса валидатора
```bash
# Проверка баланса
./build/nuahd query bank balances $(./build/nuahd keys show validator --keyring-backend test -a)
```

## 🌐 Внешний доступ

После успешного запуска нода будет доступна по следующим эндпоинтам:

- **RPC**: `http://your-server-ip:26657`
- **REST API**: `http://your-server-ip:1317`
- **gRPC**: `your-server-ip:9090`

### Быстрая проверка доступности
```bash
# Замените YOUR_SERVER_IP на реальный IP адрес
curl -s http://YOUR_SERVER_IP:26657/status | jq '.result.sync_info.latest_block_height'
```

## 🛠 Управление нодой

### Остановка ноды
```bash
# Найти процесс
ps aux | grep nuahd

# Остановить процесс (замените PID на реальный)
kill -TERM <PID>
```

### Перезапуск ноды
```bash
# Остановка
pkill nuahd

# Запуск
nohup ./build/nuahd start --rpc.laddr=tcp://0.0.0.0:26657 > nuahd.log 2>&1 &
```

### Просмотр информации о ключах
```bash
# Список всех ключей
./build/nuahd keys list --keyring-backend test

# Информация о конкретном ключе
./build/nuahd keys show validator --keyring-backend test
```

## 📊 Мониторинг

### Системные ресурсы
```bash
# Использование CPU и памяти
top -p $(pgrep nuahd)

# Использование диска
df -h
du -sh ~/.nuahd/
```

### Сетевая активность
```bash
# Количество подключенных пиров
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# Информация о пирах
curl -s http://localhost:26657/net_info | jq '.result.peers'
```

## 🔐 Безопасность

### Важные файлы для резервного копирования
```bash
# Приватный ключ валидатора
~/.nuahd/config/priv_validator_key.json

# Состояние валидатора
~/.nuahd/data/priv_validator_state.json

# Конфигурация ноды
~/.nuahd/config/config.toml
~/.nuahd/config/app.toml
```

### Создание резервной копии
```bash
# Создание архива важных файлов
tar -czf nuahd_backup_$(date +%Y%m%d).tar.gz ~/.nuahd/config/ ~/.nuahd/keyring-test/
```

## 🚨 Устранение неполадок

### Нода не запускается
1. Проверьте логи: `tail -100 nuahd.log`
2. Убедитесь, что порты свободны: `netstat -tulpn | grep :26657`
3. Проверьте права доступа к файлам конфигурации

### Нода не синхронизируется
1. Проверьте подключение к интернету
2. Убедитесь, что файрвол не блокирует исходящие соединения
3. Проверьте корректность genesis.json

### Ошибки валидации
1. Проверьте корректность chain-id
2. Убедитесь, что генезис файл валиден: `./build/nuahd validate-genesis`

## 📝 Полезные команды

```bash
# Информация о ноде
./build/nuahd status

# Информация о версии
./build/nuahd version

# Помощь по командам
./build/nuahd --help

# Экспорт состояния
./build/nuahd export

# Сброс данных ноды (ОСТОРОЖНО!)
./build/nuahd unsafe-reset-all
```

## 🎯 Результат развертывания

После успешного выполнения всех шагов у вас будет:

✅ Работающая нода блокчейна NUAH
✅ Валидатор с 5,000,000 NUAH токенами в стейке
✅ Доступные RPC, REST API и gRPC эндпоинты
✅ Автоматическое создание блоков
✅ Готовность к подключению других нод

**Chain ID**: `nuahchain-1`
**Токен**: `unuah` (1 NUAH = 1,000,000 unuah)
**Начальный баланс валидатора**: 10,000,000 NUAH
**Стейк валидатора**: 5,000,000 NUAH

---

*Эта инструкция основана на реальном развертывании ноды блокчейна NUAH. Сохраните мнемоническую фразу валидатора в безопасном месте!*
