## Сравнение с «старым» Solana API

**Что было в старом API** (`API_OLD`):
- **Аутентификация**: регистрация/логин, refresh token, logout/all, сброс пароля, Farcaster-логин, авторизация через кошелёк, привязка Telegram, просмотр сессий.
- **Профиль пользователя**: `/users/me`, обновление ника, загрузка аватарки, балансы, список токенов, активность/поинты, реферальные ссылки.
- **Рефералка**: несколько `claim`-эндпоинтов, получение referral payload.
- **Токены**: создание, creator-buy/skip, проверка имени, маркетплейс, поиск, держатели, история сделок, OHLC, supply, статус создания.
- **Торговля / котировки**: buy/sell, quotes, своп SOL↔NDOLLAR.
- **Аналитика**: графики, статистика рынка, исторические данные.
- **Автоторговля** (`TRADING-ROUTES`): боты (CRUD), стратегии, бэктесты, performance-дашборды, глобальный статус.

**Что есть в новом Cosmos-сервере**
- **Эндпоинты**:
  - `/health`, `/health/db`
  - `/api/auth/{register|login|telegram}`, `/api/auth/me`
  - `/api/tokens/{create|buy|sell}` (x/usertokens)
  - `/api/assets/{ensure|buy|sell|margin/open|margin/close}` (x/assets + x/leverage)
  - `/api/stablecoin/{buy-ndollar|sell-ndollar}` (x/stablecoin)
  - `/api/exchange/tokens` (x/exchange)
  - `/api/tx/{hash}`
  - `/api/admin/{transactions, transactions/ws, stats}` (мониторинг)
- **Инфраструктура**:
  - Клиенты для модулей: assets, leverage, stablecoin, exchange, usertokens, tokenfactory.
  - NDOLLAR модули настроены (скрипты создают токен до старта цепи, Postgres + миграции поднимаются автоматически).
  - Асинхронный трекер транзакций с WebSocket + резервный polling.
  - Админ-монитор: WebSocket слушает `tm.event='Tx'`, пишет события в БД; обогащение (grpc GetTx) можно включить по требованию.
  - Тестовые скрипты (`test_assets.sh`, `test_monitor.sh`).

## Где мы отстаём от старого API

| Зона | Было на Solana | Статус в новом сервере | Что делать |
| --- | --- | --- | --- |
| Auth/сессии | Refresh, logout/all, reset, Farcaster, wallet-signature | Только register/login/Telegram | Добавить refresh, logout, восстановление пароля, кошельковую auth при необходимости |
| Сессии/устройства | `/auth/sessions`, `/auth/logout-all` | Нет | Нужен storage для сессий |
| Профиль/балансы | `/users/me`, info, username, image, balances, tokens | Только `/api/auth/me` | Добавить модель профиля, загрузку файлов, агрегацию балансов |
| Маркетплейс токенов | Поиск, листинг, держатели, OHLC, history, tx history | Нет | Требуется индексатор по событиям блокчейна + эндпоинты |
| Котировки | `/quote/trade`, `/quote/swap` | Нет | Нужен сервис котировок (oracle/ценник) |
| NDollar ↔ SOL | Swap SOL⇄NDollar | Космос-версия пока только mint/burn N$ | Нужен мост или определение эквивалента SOL |
| Рефералы/поинты | Много эндпоинтов | Нет | Нужны таблицы и логика |
| Автоторговля | Боты, стратегии, бэктест | Нет | Надо решать, делаем ли Cosmos-модуль или off-chain сервис |
| Creator controls | creator buy/skip, candlestick interval | Нет | Определить поддержку в модулях/сервере |
| Wallet auth | `/auth/wallet/start/verify` | Нет | Настроить проверку подписей Cosmos |
| Загрузка аватарок | `/users/me/upload-image` | Нет | Хранилище файлов |
| Marketplace аналитика | Market/holders/ohlc | Нет | Нужен индексатор/аналитика |

## Состояние блокчейн-модулей

| Модуль | Реализация на цепи | Сервер | Недостающее |
| --- | --- | --- | --- |
| `x/assets` | ensure/buy/sell | Эндпоинты есть | Требуются БД-модели рынка, аналитика |
| `x/leverage` | margin open/close | Подключено через `/api/assets/margin/*` | Нужно хранить позиции, отдавать PnL |
| `x/stablecoin` | NDollar mint/burn | `/api/stablecoin/*` | Балансы/котировки для клиентов |
| `x/exchange` | swap токенов → unuah | `/api/exchange/tokens` | Добавить котировки, возможно обратный swap |
| `x/usertokens` | выпуск токенов | Эндпоинты create/buy/sell | Нет листинга/holders |
| `x/usdoracle` / `x/pegkeeper` | ценовой oracle и стабилизация | Есть на цепи, пока не отдаются наружу | Имеет смысл добавить сервис чтения oracle |
| Мониторинг | WebSocket + БД | Эндпоинты admin | Обогащение отключено, можно включать по требованию |

## Рекомендации

1. **Выровнять auth**: refresh/logout/password reset + альт-логины.
2. **Пользовательские данные**: balances, профили, referrals, activity.
3. **Маркетплейс/аналитика**: индексировать события из монитора, строить поиск/графики.
4. **Quotes/Oracle API**: использовать x/oracle или отдельный ценник.
5. **NDollar-bridge**: решить, как заменить SOL в новой архитектуре (IBC/другой denom).
6. **Реферальная система**: перенести бизнес-логику и таблицы.
7. **Автоторговля**: определить архитектуру (Cosmos module vs off-chain) и реализовать API.
8. **Документация**: синхронизировать новые эндпоинты с клиентскими ожиданиями.

Такое резюме показывает, что ядро (assets/leverage/stablecoin) уже на Cosmos и отдается через API, но весь «обвес» старого продукта (auth, профиль, маркетинг, аналитика, боты) нужно постепенно переносить или переосмысливать в новой архитектуре.
