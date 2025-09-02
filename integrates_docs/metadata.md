# Установка метаданных для токенов в Osmosis

## Обзор

Документация описывает процесс установки метаданных для токенов, созданных через модуль `tokenfactory` в Osmosis. Метаданные включают человекочитаемое название, символ, описание и единицы измерения токена.

## Проблема

При попытке использовать команду `set-denom-metadata` возникала ошибка:

```bash
Error: accepts 0 arg(s), received 1
```

Это происходило из-за того, что CLI система не могла правильно парсить сложную структуру `cosmos.bank.v1beta1.Metadata`.

## Решение

### 1. Добавление CustomFieldParser

В файле `x/tokenfactory/client/cli/tx.go` добавлен кастомный парсер для поля `Metadata`:

```go
func NewMsgSetDenomMetadata() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgSetDenomMetadata](&osmocli.TxCliDesc{
		Use:     "set-denom-metadata",
		Short:   "overwriting of the denom metadata in the bank module.",
		NumArgs: 1,
		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{
			"Metadata": parseMetadataField,
		},
	})
}

// parseMetadataField parses the metadata field from JSON string to banktypes.Metadata
func parseMetadataField(arg string, _ *pflag.FlagSet) (any, bool, error) {
	var metadata banktypes.Metadata
	err := json.Unmarshal([]byte(arg), &metadata)
	if err != nil {
		return nil, true, err
	}
	return metadata, true, nil
}
```

### 2. Исправление логики подсчета аргументов

В файле `osmoutils/osmocli/tx_cmd_wrap.go` исправлена логика подсчета аргументов:

```go
func BuildTxCli[M sdk.Msg](desc *TxCliDesc) *cobra.Command {
	desc.TxSignerFieldName = strings.ToLower(desc.TxSignerFieldName)
	if desc.NumArgs == 0 {
		// NumArgs = NumFields - 1, since 1 field is from the msg
		// CustomFieldParsers don't reduce NumArgs since they parse from arguments, not flags
		desc.NumArgs = ParseNumFields[M]() - 1 - len(desc.CustomFlagOverrides)
	}
	// ... rest of the function
}
```

## Структура метаданных

Метаданные токена должны соответствовать структуре `cosmos.bank.v1beta1.Metadata`:

```json
{
  "description": "Описание токена",
  "denom_units": [
    {
      "denom": "factory/creator_address/subdenom",
      "exponent": 0,
      "aliases": ["uroma"]
    },
    {
      "denom": "roma",
      "exponent": 6
    }
  ],
  "base": "factory/creator_address/subdenom",
  "display": "roma",
  "name": "Название токена",
  "symbol": "ROMA"
}
```

### Поля метаданных:

- **description**: Описание токена
- **denom_units**: Массив единиц измерения
  - **denom**: Полное название единицы
  - **exponent**: Показатель степени (0 для базовой единицы)
  - **aliases**: Альтернативные названия
- **base**: Базовая единица (должна совпадать с denom токена)
- **display**: Единица для отображения пользователю
- **name**: Человекочитаемое название
- **symbol**: Символ токена

## Пошаговая инструкция

### 1. Получение адреса создателя

```bash
CREATOR=$(./build/osmosisd keys show alice -a --keyring-backend test)
echo "CREATOR: $CREATOR"
```

### 2. Создание denom

```bash
DENOM="factory/$CREATOR/roma"
echo "DENOM: $DENOM"
```

### 3. Создание файла метаданных

```bash
cat > metadata.json << 'EOF'
{
  "description": "Roma Token",
  "denom_units": [
    { "denom": "FACTORY_PLACEHOLDER", "exponent": 0, "aliases": ["uroma"] },
    { "denom": "roma", "exponent": 6 }
  ],
  "base": "FACTORY_PLACEHOLDER",
  "display": "roma",
  "name": "Roma Token",
  "symbol": "ROMA"
}
EOF
```

### 4. Замена placeholder на реальный denom

```bash
sed -i '' "s|FACTORY_PLACEHOLDER|$DENOM|g" metadata.json
cat metadata.json
```

### 5. Выполнение команды

```bash
./build/osmosisd tx tokenfactory set-denom-metadata "$(cat metadata.json)" \
  --from alice \
  --keyring-backend test \
  --chain-id localosmosis \
  --gas auto \
  --gas-adjustment 1.5 \
  --fees 2500stake \
  -y
```

## Проверка результата

### 1. Проверка метаданных

```bash
./build/osmosisd query bank denom-metadata factory/creator_address/roma --chain-id localosmosis
```

### 2. Проверка списка токенов создателя

```bash
./build/osmosisd query tokenfactory denoms-from-creator creator_address --chain-id localosmosis
```

### 3. Проверка баланса

```bash
./build/osmosisd query bank balances creator_address --chain-id localosmosis
```

## Пример успешного выполнения

```bash
gas estimate: 103080
code: 0
codespace: ""
data: ""
events: []
gas_used: "0"
gas_wanted: "0"
height: "0"
info: ""
logs: []
raw_log: ""
timestamp: ""
tx: null
txhash: 50039A56998D63BCA190D85BD2CEBFBEC98702F7AC810D737F48A9D5603B5FA2
```

## Важные замечания

1. **Fee токены**: Используйте правильный fee токен для вашего чейна (например, `stake` для localosmosis)
2. **Chain ID**: Всегда указывайте правильный `--chain-id`
3. **Права доступа**: Только администратор токена может изменять его метаданные
4. **Формат JSON**: Метаданные должны быть в валидном JSON формате

## Результат

После успешного выполнения команды:
- Токен будет отображаться с человекочитаемым названием
- В кошельках и интерфейсах будет показываться символ `ROMA` вместо длинной строки
- Метаданные сохранятся в блокчейне и будут доступны для всех приложений

## Технические детали

- **Модуль**: `x/tokenfactory`
- **Команда**: `set-denom-metadata`
- **Тип сообщения**: `MsgSetDenomMetadata`
- **Структура**: `cosmos.bank.v1beta1.Metadata`
- **Парсер**: Кастомный JSON парсер для CLI
