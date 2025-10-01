#!/bin/bash

# Скрипт для создания специальных кошельков и обновления параметров модуля usertoken

set -e

CHAIN_ID="nuahchain"
KEYRING_BACKEND="test"
NODE="tcp://localhost:26657"
BINARY="./build/nuahd"

# Функция для создания кошелька, если он не существует
create_wallet_if_not_exists() {
    local wallet_name=$1
    
    # Проверяем, существует ли кошелек
    if $BINARY keys show $wallet_name --keyring-backend $KEYRING_BACKEND >/dev/null 2>&1; then
        echo "Кошелек $wallet_name уже существует"
        address=$($BINARY keys show $wallet_name --keyring-backend $KEYRING_BACKEND -a)
    else
        echo "Создаем кошелек $wallet_name..."
        output=$($BINARY keys add $wallet_name --keyring-backend $KEYRING_BACKEND --output json)
        address=$(echo "$output" | jq -r '.address')
        echo "Создан кошелек $wallet_name с адресом: $address"
    fi
    
    echo $address
}

echo "=== Создание специальных кошельков для модуля usertoken ==="

# Создаем кошельки
AI_CEO_WALLET=$(create_wallet_if_not_exists "ai-ceo-wallet")
PLATFORM_WALLET=$(create_wallet_if_not_exists "platform-wallet")
REFERRAL_WALLET=$(create_wallet_if_not_exists "referral-wallet")

echo ""
echo "Созданные кошельки:"
echo "AI CEO Wallet: $AI_CEO_WALLET"
echo "Platform Wallet: $PLATFORM_WALLET"
echo "Referral Wallet: $REFERRAL_WALLET"

echo ""
echo "=== Обновление параметров модуля usertoken ==="

# Получаем адрес валидатора для authority
VALIDATOR_ADDRESS=$($BINARY keys show validator --keyring-backend $KEYRING_BACKEND -a)

# Обновляем параметры модуля с помощью новой CLI команды
$BINARY tx usertoken update-params \
    "1000000" \
    "100000000" \
    "0.000001" \
    "10.0" \
    "1000000000" \
    "1.0" \
    "$AI_CEO_WALLET" \
    "$REFERRAL_WALLET" \
    "$PLATFORM_WALLET" \
    --from validator \
    --keyring-backend $KEYRING_BACKEND \
    --chain-id $CHAIN_ID \
    --node $NODE \
    --gas auto \
    --gas-adjustment 1.5 \
    --yes

echo "Параметры модуля usertoken успешно обновлены!"
echo ""
echo "Конфигурация завершена. Специальные кошельки настроены:"
echo "- AI CEO Wallet (40M токенов): $AI_CEO_WALLET"
echo "- Platform Wallet (10M токенов): $PLATFORM_WALLET"  
echo "- Referral Wallet (10M токенов): $REFERRAL_WALLET"