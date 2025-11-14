# Анализ: Как пользователь может купить unuah и NDOLLAR

## 📋 Текущее состояние

### 1. **Существующие модули**

#### ✅ `x/exchange` - Обмен криптовалют на unuah (ГОТОВ!)
**Уже реализован и работает!**

- **Назначение**: Обмен популярных криптовалют → unuah (N$)
- **Поддерживаемые токены**:
  - ETH (ERC-20 через IBC)
  - BTC (через IBC)
  - USDC (через IBC)
  - USDT (через IBC)
  - ATOM (через IBC)
  - OSMO (нативный)
  - SOL (через IBC)

**Как работает:**
```
Крипта (ETH/BTC/USDC...) → x/exchange → unuah (mintится 1:1 с USD)
```

**Ключевые features:**
- ✅ Dual Oracle (USD Oracle + TWAP для валидации цен)
- ✅ Лимиты: $10 min, $100k max за транзакцию
- ✅ Daily limit: $1M на адрес
- ✅ Комиссия: 0.1% (конфигурируемая)
- ✅ Anti-whale защита
- ✅ Governance управление параметрами

**Пример использования:**
```bash
# Обменять 100 OSMO на unuah
nuahd tx exchange exchange-tokens 100000000uosmo <recipient_address> --from mykey

# Обменять 1000 USDC на unuah
nuahd tx exchange exchange-tokens 1000000000ibc/D189335... <recipient> --from mykey
```

---

#### ✅ `x/stablecoin` - Конвертация unuah ↔ NDOLLAR (1:1)
**Уже реализован!**

- **BuyNDollar**: unuah → NDOLLAR (1:1)
- **SellNDollar**: NDOLLAR → unuah (1:1)
- Интегрирован в сервер: `/api/stablecoin/buy-ndollar`, `/api/stablecoin/sell-ndollar`

---

#### ✅ `x/mint` - Автоматический минтинг unuah через эпохи
**Инфляционная модель (как в Osmosis):**

- Минтит unuah каждую эпоху (по умолчанию: daily)
- Распределение:
  - 25% → Staking rewards
  - 45% → Pool incentives
  - 25% → Developer rewards
  - 5% → Community pool
- Reduction factor: 66.66% каждый год (3 года = 1 период)
- Finite supply

---

## 🎯 ЧТО УЖЕ ЕСТЬ (работает сейчас)

### Путь 1: Крипта → unuah
```
1. Пользователь владеет: ETH/BTC/USDC/USDT/ATOM/OSMO/SOL
2. Вызывает: nuahd tx exchange exchange-tokens <amount> <recipient>
3. Результат: Получает unuah (mintится на основе USD цены оракула)
```

### Путь 2: unuah → NDOLLAR
```
1. Пользователь владеет: unuah
2. Вызывает: nuahd tx stablecoin buy-ndollar <amount>
   ИЛИ через API: POST /api/stablecoin/buy-ndollar {"amount": "1000000"}
3. Результат: Получает NDOLLAR (1:1)
```

### Путь 3: NDOLLAR → unuah
```
1. Пользователь владеет: NDOLLAR
2. Вызывает: nuahd tx stablecoin sell-ndollar <amount>
   ИЛИ через API: POST /api/stablecoin/sell-ndollar {"amount": "1000000"}
3. Результат: Получает unuah обратно (1:1)
```

---

## 🚀 ЧТО НАДО ДОБАВИТЬ

### Для покупки за ФИАТ (на доверии, без пулов ликвидности)

#### Вариант А: Расширение `x/exchange` для фиата

**Идея**: Добавить в `x/exchange` поддержку фиатных валют

**Архитектура:**
```
┌─────────────┐      API/Web Interface     ┌──────────────┐
│  User       │─────────────────────────────▶│   Admin      │
│  (Buyer)    │   "Хочу купить $100 unuah"  │  (Validator) │
└─────────────┘                              └──────────────┘
                                                     │
                                                     ▼
                                             ┌──────────────┐
                                             │ Проверяет    │
                                             │ оплату фиат  │
                                             └──────────────┘
                                                     │
                                                     ▼
                                             ┌──────────────┐
                                             │ Минтит unuah │
                                             │ напрямую     │
                                             └──────────────┘
```

**Что добавить:**

1. **Новое сообщение в `x/exchange`:**
```protobuf
message MsgRequestFiatPurchase {
  string buyer = 1;
  string amount_usd = 2;  // Сумма в USD
  string payment_method = 3;  // "bank_transfer", "card", "paypal" и т.д.
  string payment_proof = 4;  // Ссылка на подтверждение оплаты
  string currency = 5;  // "USD", "EUR", "RUB" и т.д.
}

message MsgApproveFiatPurchase {
  string admin = 1;
  string request_id = 2;
  bool approved = 3;
  string tx_reference = 4;  // Банковская ссылка для аудита
}
```

2. **Новая структура в state:**
```go
type FiatPurchaseRequest struct {
  ID            string
  Buyer         string
  AmountUSD     sdk.Dec
  PaymentMethod string
  PaymentProof  string
  Currency      string
  Status        string  // "PENDING", "APPROVED", "REJECTED"
  CreatedAt     time.Time
  ProcessedAt   time.Time
  ProcessedBy   string  // Admin address
}
```

3. **Workflow:**
```
1. Пользователь → POST /api/fiat/request-purchase
   {
     "amount_usd": "100",
     "payment_method": "bank_transfer",
     "payment_proof": "receipt_123.pdf",
     "currency": "USD"
   }

2. Система → Создает FiatPurchaseRequest (status: PENDING)

3. Админ → Проверяет банковский перевод вручную

4. Админ → POST /api/fiat/approve-purchase
   {
     "request_id": "req_123",
     "approved": true,
     "tx_reference": "BANK_TXN_456"
   }

5. Blockchain → Минтит unuah пользователю (1:1 с USD)
```

---

#### Вариант Б: Новый модуль `x/fiatonramp`

**Более структурированный подход с отдельным модулем:**

```go
// x/fiatonramp/types/tx.proto
message MsgCreateFiatOrder {
  string buyer = 1;
  string amount_fiat = 2;
  string fiat_currency = 3;  // USD, EUR, RUB
  string payment_method = 4;
  string payment_details = 5;
}

message MsgFulfillFiatOrder {
  string admin = 1;
  uint64 order_id = 2;
  string proof_of_payment = 3;
}
```

**Features:**
- Escrow system (опционально)
- Multi-currency support
- Payment gateway integration (Stripe, PayPal)
- KYC/AML compliance helpers
- Audit trail

---

### Для интеграции с Payment Gateways

#### Вариант В: Stripe/PayPal Integration (автоматизация)

**Server-side компонент:**

```go
// server/payment/stripe.go
func HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
  // 1. Валидация webhook от Stripe
  // 2. Проверка payment_intent.succeeded
  // 3. Автоматический минт unuah через blockchain
}
```

**Workflow:**
```
1. User → Фронтенд с Stripe Checkout
2. Stripe → Обрабатывает платеж (USD/EUR/etc)
3. Stripe → Webhook → Наш сервер
4. Server → Blockchain.MintUnuah(user_address, amount)
5. User → Получает unuah в кошелек
```

---

## 📊 Сравнение вариантов

| Вариант | Сложность | Автоматизация | Trust Model | Время внедрения |
|---------|-----------|---------------|-------------|----------------|
| **А: Расширение x/exchange** | Низкая | Низкая (ручная проверка) | Полное доверие к админу | 1-2 недели |
| **Б: x/fiatonramp** | Средняя | Средняя | Structured trust | 3-4 недели |
| **В: Payment Gateway** | Высокая | Высокая (автомат) | Trust to gateway | 4-6 недель |

---

## 💡 РЕКОМЕНДАЦИЯ

### Phase 1: MVP (на доверии) - **РЕКОМЕНДУЮ НАЧАТЬ С ЭТОГО**

**Что делать:**

1. ✅ Использовать существующий `x/exchange` как foundation
2. ✅ Добавить простой web-интерфейс:
   ```
   - Форма: "Купить unuah за фиат"
   - Банковские реквизиты для перевода
   - Upload подтверждения оплаты
   ```
3. ✅ Админка для ручного апрува:
   ```
   - Список запросов на покупку
   - Проверка банковского перевода
   - Кнопка "Approve" → минтит unuah
   ```

**Преимущества:**
- 🚀 Быстрый запуск (1-2 недели)
- 💰 Минимальные затраты
- 🔒 Полный контроль
- 📈 Можно тестировать спрос

**Недостатки:**
- ⏱️ Ручная обработка (медленно)
- 👤 Требует доверия к админу
- 📊 Не масштабируется

---

### Phase 2: Automation (через 3-6 месяцев)

**После валидации спроса:**

1. Интеграция Stripe/PayPal для автоматических платежей
2. KYC/AML компл

иенс (если требуется по регуляциям)
3. Автоматический минтинг через webhooks
4. Возможно, создание DAO для децентрализации апрува

---

## 🔧 ТЕХНИЧЕСКИЕ ДЕТАЛИ

### Для Варианта А (MVP):

**1. Создать новый файл:**
```go
// server/fiat/models.go
type FiatPurchaseRequest struct {
    ID            int64
    UserID        int64
    AmountUSD     string
    PaymentMethod string
    PaymentProof  string
    Status        string  // PENDING, APPROVED, REJECTED
    CreatedAt     time.Time
    ProcessedAt   *time.Time
}
```

**2. API Endpoints:**
```
POST   /api/fiat/request-purchase    - Создать запрос
GET    /api/fiat/my-requests          - Мои запросы
GET    /api/admin/fiat/pending        - (Admin) Pending requests
POST   /api/admin/fiat/approve/{id}   - (Admin) Approve & mint
```

**3. Admin Workflow:**
```
1. Админ проверяет банковский счет
2. Находит перевод от пользователя
3. Проверяет amount & details
4. Вызывает blockchain: MintCoins(user_address, amount)
5. Обновляет Status → APPROVED
```

**4. Blockchain Integration:**
```go
// Использовать существующий x/exchange или x/tokenfactory
func (k Keeper) MintUnuahToUser(ctx sdk.Context, recipient sdk.AccAddress, amount sdkmath.Int) error {
    coin := sdk.NewCoin("unuah", amount)

    // Mint to module
    if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
        return err
    }

    // Send to user
    return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(coin))
}
```

---

## 🎯 ИТОГОВАЯ АРХИТЕКТУРА

```
┌─────────────────────────────────────────────────────────────┐
│                     PURCHASE FLOW                            │
└─────────────────────────────────────────────────────────────┘

[User] ─── Фиат (USD/EUR) ──▶ [Bank Transfer] ──▶ [Admin Checks]
                                                           │
                                                           ▼
                                                    [Blockchain]
                                                           │
                                                           ▼
                                                    [Mint unuah]
                                                           │
                                                           ▼
                                                      [User Wallet]

┌─────────────────────────────────────────────────────────────┐
│            EXISTING: CRYPTO → unuah (WORKING!)              │
└─────────────────────────────────────────────────────────────┘

[User] ─── ETH/BTC/USDC ──▶ [x/exchange] ──▶ [Mint unuah]
                                                     │
                                                     ▼
                                               [User Wallet]

┌─────────────────────────────────────────────────────────────┐
│          unuah ↔ NDOLLAR (1:1) (WORKING!)                   │
└─────────────────────────────────────────────────────────────┘

[User] ─── unuah ──▶ [x/stablecoin] ──▶ [NDOLLAR]
[User] ─── NDOLLAR ──▶ [x/stablecoin] ──▶ [unuah]
```

---

## 📝 ВЫВОДЫ

### ✅ ЧТО УЖЕ РАБОТАЕТ:

1. ✅ **Покупка unuah за криптовалюту** (`x/exchange`)
   - ETH, BTC, USDC, USDT, ATOM, OSMO, SOL → unuah

2. ✅ **Конвертация unuah ↔ NDOLLAR** (`x/stablecoin`)
   - 1:1 обмен между unuah и NDOLLAR

3. ✅ **Автоматический минтинг** (`x/mint`)
   - Инфляционная модель для стейкинга/пулов

### ⚠️ ЧТО НАДО ДОБАВИТЬ:

1. 🔨 **Покупка unuah за ФИАТ**
   - Рекомендация: MVP на доверии (1-2 недели)
   - Затем: автоматизация через Stripe/PayPal

### 🎯 PLAN OF ACTION:

**Week 1-2: MVP Fiat On-Ramp**
- [ ] Создать `server/fiat` модуль
- [ ] API endpoints для запросов на покупку
- [ ] Admin panel для апрува
- [ ] Blockchain integration (mint)

**Week 3-4: Testing & Launch**
- [ ] Тестирование с реальными банковскими переводами
- [ ] Документация для пользователей
- [ ] Публичный запуск

**Month 2-3: Automation (если есть спрос)**
- [ ] Stripe/PayPal integration
- [ ] Автоматический минтинг через webhooks
- [ ] KYC/AML если требуется

---

## 🔐 БЕЗОПАСНОСТЬ

**Важно для MVP на доверии:**

1. ✅ Ограничить права минтинга только админу
2. ✅ Логировать все операции (audit trail)
3. ✅ Лимиты на покупку (например, $10k max в день)
4. ✅ KYC для крупных сумм (>$1000)
5. ✅ Мультисиг для админа (2-of-3 или 3-of-5)

---

**Готово к имплементации!** 🚀

Хотите начать с MVP фиата на доверии?

