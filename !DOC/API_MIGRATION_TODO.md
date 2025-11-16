# 📋 TODO: Миграция функционала из старого Solana API

> **Цель**: Перенести недостающий функционал из старого Solana API в новый Cosmos-сервер
> **Статус**: В работе
> **Последнее обновление**: 2025-11-14

---

## 🔐 Аутентификация и сессии

### Высокий приоритет
- [x] **Refresh tokens** - Реализовать `/api/auth/refresh` для обновления access token ✅
- [x] **Logout** - Реализовать `/api/auth/logout` (инвалидация текущего токена) ✅
- [x] **Logout all** - Реализовать `/api/auth/logout-all` (инвалидация всех сессий пользователя) ✅
- [x] **Password reset** - Реализовать `/api/auth/web/forgot-password` и `/api/auth/web/reset-password` ✅
- [x] **Sessions management** - Реализовать `/api/auth/sessions` (список активных сессий) ✅

### Средний приоритет
- [ ] **Farcaster login** - Реализовать `/api/auth/farcaster/login` (если требуется)
- [ ] **Wallet authentication** - Реализовать `/api/auth/wallet/start` и `/api/auth/wallet/verify` (Cosmos signature verification)

---

## 👤 Профиль пользователя и данные

### Высокий приоритет
- [x] **User profile extended** - Расширить `/api/auth/me` до `/api/users/me` с полной информацией ✅
- [x] **User info summary** - Реализовать `/api/users/me/info` (краткая сводка профиля) ✅
- [x] **Username update** - Реализовать `PATCH /api/users/username` для обновления username ✅
- [x] **Image upload** - Реализовать `POST /api/users/me/upload-image` (загрузка аватарок) ✅

### Средний приоритет
- [ ] **User balances** - Реализовать `/api/users/balances-db` (агрегация балансов из блокчейна)
- [x] **User tokens list** - Реализовать `/api/users/me/tokens` (список токенов пользователя) ✅
- [ ] **Activity points** - Реализовать `/api/users/me/activity/points` (система активности/поинтов)

---

## 🪙 Токены и маркетплейс

### Высокий приоритет
- [x] **Token marketplace** - Реализовать `GET /api/tokens/market` (листинг всех токенов) ✅
- [x] **Token search** - Реализовать `GET /api/tokens/search?query={query}` (поиск по имени/символу) ✅
- [x] **Token details** - Реализовать `GET /api/tokens/{denom}/details` (детальная информация) ✅
- [ ] **Token holders** - Реализовать `GET /api/tokens/{mintAddress}/holders` (список держателей)
- [ ] **Token transaction history** - Реализовать `GET /api/tokens/{mintAddress}/tx-history` (история транзакций)

### Средний приоритет
- [ ] **Token price history** - Реализовать `GET /api/tokens/{mintAddress}/history` (история цен)
- [ ] **OHLC data** - Реализовать `GET /api/tokens/{mintAddress}/ohlc` (свечи для графиков)
- [ ] **Available supply** - Реализовать `GET /api/tokens/{mintAddress}/available-supply`
- [ ] **Token creation status** - Реализовать `GET /api/tokens/creation-status` и `/api/tokens/{mintAddress}/creation-status`
- [ ] **Check token name** - Реализовать `GET /api/tokens/check-name?name={name}` (проверка доступности имени)

### Низкий приоритет
- [ ] **Creator controls** - Реализовать `/api/tokens/creator-buy` и `/api/tokens/skip-creator-buy`
- [ ] **Candlestick interval** - Реализовать `POST /api/tokens/{tokenMintAddress}/candlestick-interval`

**Примечание**: Для реализации маркетплейса потребуется:
- Индексация событий блокчейна (расширить текущий монитор)
- Таблицы БД для токенов, держателей, истории цен
- Агрегация данных из `blockchain_transactions`

---

## 💱 Котировки и торговля

### Высокий приоритет
- [ ] **Trade quotes** - Реализовать `GET /api/quote/trade?tokenMintAddress={}&amount={}&operation={buy|sell}&inputType={}`
- [ ] **Swap quotes** - Реализовать `GET /api/quote/swap-sol?solAmount={}` и `/api/n-dollar/quote-reverse?ndollarAmount={}`

**Примечание**: Требуется интеграция с `x/oracle` или внешним ценовым сервисом для получения актуальных цен

---

## 💵 NDollar и свопы

### Средний приоритет
- [ ] **SOL ↔ NDollar bridge** - Определить архитектуру замены SOL в Cosmos-контексте
  - Варианты: IBC bridge, симулированный denom, или другой подход
- [ ] **Swap endpoints** - Реализовать `/api/n-dollar/swap-sol` и `/api/n-dollar/swap-ndollar` (если требуется)

**Примечание**: Текущая реализация использует mint/burn через `x/stablecoin`, но нет прямого эквивалента SOL

---

## 🎁 Реферальная система

### Средний приоритет
- [ ] **Referral links** - Реализовать `GET /api/referrals/me/link` (генерация реферальной ссылки)
- [ ] **Referral users** - Реализовать `GET /api/referrals/me/users` (список приглашенных пользователей)
- [ ] **Referral links list** - Реализовать `GET /api/users/me/referral-links` (все ссылки пользователя)
- [ ] **Claim referral** - Реализовать `POST /api/referrals/claim` и `/api/referrals/user/claim`
- [ ] **Referral payload** - Реализовать `GET /api/tokens/{mintAddress}/referral-payload`

**Примечание**: Требуется:
- Таблицы БД для рефералов (`referrals`, `referral_links`)
- Бизнес-логика для начисления наград
- Интеграция с токенами/активами

---

## 🤖 Автоторговля (Autotrade)

### Низкий приоритет (требует архитектурного решения)
- [ ] **Определить архитектуру** - Cosmos module vs off-chain сервис
- [ ] **Trading bots CRUD** - Реализовать:
  - `POST /api/autotrade/bots` (создание)
  - `GET /api/autotrade/bots` (список)
  - `GET /api/autotrade/bots/{id}` (детали)
  - `PUT /api/autotrade/bots/{id}` (обновление)
  - `DELETE /api/autotrade/bots/{id}` (удаление)
  - `POST /api/autotrade/bots/{id}/start` и `/stop` (управление)
- [ ] **Trading strategies** - Реализовать:
  - `GET /api/autotrade/strategies` (список стратегий)
  - `GET /api/autotrade/strategies/{name}` (детали стратегии)
  - `POST /api/autotrade/strategies/{name}/backtest` (бэктест)
- [ ] **Market analysis** - Реализовать:
  - `GET /api/autotrade/market/analysis/{token}`
  - `GET /api/autotrade/market/candlesticks/{token}`
  - `GET /api/autotrade/market/conditions/{token}`
- [ ] **Performance tracking** - Реализовать:
  - `GET /api/autotrade/performance/overview`
  - `GET /api/autotrade/performance/bots/{id}`
  - `GET /api/autotrade/performance/history`
- [ ] **System status** - Реализовать:
  - `GET /api/autotrade/get-status`
  - `POST /api/autotrade/set-status`
  - `GET /api/autotrade/details`

**Примечание**: Это большая система, требующая:
- Хранение ботов и стратегий в БД
- Интеграция с блокчейном для выполнения сделок
- Система бэктестинга
- Performance метрики

---

## 🔗 Дополнительные интеграции

### Низкий приоритет
- [ ] **Telegram link token** - Реализовать `POST /api/users/me/link-telegram-token` (если требуется)
- [ ] **Batch token names** - Реализовать `POST /api/tokens/batch-names` (генерация имен токенов)

---

## 📊 Инфраструктура и улучшения

### Высокий приоритет
- [ ] **Индексатор маркетплейса** - Расширить `blockchain_transactions` монитор для индексации:
  - Создание токенов
  - Торговые операции
  - Изменения цен
  - Держатели токенов
- [ ] **База данных для маркетплейса** - Создать таблицы:
  - `tokens` (метаданные токенов)
  - `token_holders` (держатели)
  - `token_prices` (история цен)
  - `token_ohlc` (свечи)
- [ ] **Oracle/Quotes сервис** - Интеграция с `x/oracle` для получения цен активов

### Средний приоритет
- [ ] **Transaction enrichment toggle** - Добавить конфигурацию для включения/выключения обогащения транзакций
- [ ] **API документация** - Обновить документацию API с новыми эндпоинтами
- [ ] **Балансы пользователей** - Агрегация балансов из блокчейна (query balances)
- [ ] **Exchange rate auto-update** - Автоматическое обновление exchange rate при запросе котировки, если токен поддерживается, но rate отсутствует (требует oracle и TWAP данные)

---

## ✅ Уже реализовано

- ✅ Базовая аутентификация (register/login/telegram)
- ✅ Создание токенов (`/api/tokens/create`)
- ✅ Торговля токенами (`/api/tokens/buy`, `/api/tokens/sell`)
- ✅ Торговля активами (`/api/assets/ensure`, `/api/assets/buy`, `/api/assets/sell`)
- ✅ Маржинальная торговля (`/api/assets/margin/open`, `/api/assets/margin/close`)
- ✅ NDollar buy/sell (`/api/stablecoin/buy-ndollar`, `/api/stablecoin/sell-ndollar`)
- ✅ Exchange токенов (`/api/exchange/tokens`)
- ✅ Статус транзакций (`/api/tx/{hash}`)
- ✅ Админ мониторинг (`/api/admin/transactions`, `/api/admin/stats`)
- ✅ WebSocket для транзакций (встроен в tracker и админ панель)

---

## 📝 Примечания

1. **Приоритеты** основаны на важности для базового функционала продукта
2. **Маркетплейс** требует значительной работы по индексации данных блокчейна
3. **Автоторговля** - самая сложная система, требует архитектурного решения
4. **Реферальная система** может быть реализована независимо от других компонентов
5. Все новые эндпоинты должны следовать текущей архитектуре сервера (Go handlers, DB models, blockchain client)

---

**Следующие шаги**: Начать с аутентификации (refresh/logout) и профиля пользователя, затем маркетплейс и котировки.

