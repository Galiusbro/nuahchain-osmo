Инструкция по установке и развертыванию Mars Protocol CosmWasm (Red Bank) на вашем блокчейне с поддержкой CosmWasm:

***

### 1. Подготовка окружения

- Убедитесь, что ваша сеть поддерживает CosmWasm.
- Необходимы: установленный Rust (nightly), cargo-make, Docker, Node.js v16, Yarn.[1][2]

```bash
# Установите Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
rustup default stable
rustup target add wasm32-unknown-unknown

# Установите cargo-make
cargo install --force cargo-make
```

***

### 2. Сборка CosmWasm-контрактов Mars Protocol

- Клонируйте репозиторий с core-контрактами:

```bash
git clone https://github.com/mars-protocol/core-contracts.git
cd core-contracts
```

- Установите зависимости:

```bash
cd scripts
yarn
yarn build
```

- Скомпилируйте все контракты:

```bash
cargo make rust-optimizer
```

Готовые `.wasm` файлы появятся в папке `/artifacts`.[1]

***

### 3. Загрузка и деплой контракта в сеть

- Загрузите контракт в вашу CosmWasm-сеть:

```bash
wasmd tx wasm store artifacts/<contract>.wasm --from <your-key> --gas auto -y
```

- Получите code_id из respone.

- Проинстанциируйте контракт:

```bash
wasmd tx wasm instantiate <code_id> '<init_message>' --from <your-key> --label "red-bank" --gas auto -y
```

- Получите адрес контракта из ответа.

***

### 4. Инициализация и интеграция

- Инициализируйте контракт init_msg с нужными активами и параметрами.
- Привяжите контракт к существующему маркету, установите ораклы.
- Для массового деплоя используйте скрипты из директории scripts.[3][1]

***

### 5. Проверка работоспособности и тестирование

- Проверьте основные функции: supply, borrow, repay, withdraw — через wasmcli или frontend.[3]
- Проведите интеграционные тесты (unit-тесты идут вместе с контрактами).

***

### 6. Управление и обновления

- Управляйте контрактом через DAO-механизмы вашей сети или Mars governance-сообщениями.
- Для обновлений используйте функцию migrate существующего CosmWasm контракта.

***

**Документация и исходники:**
- Контракты: https://github.com/mars-protocol/core-contracts
- Mars Protocol docs: https://docs.marsprotocol.io/smart-contracts/red-bank

***

Эта инструкция универсальна для внедрения на любой Cosmos-chain c интеграцией CosmWasm.[4][1][3]

[1](https://github.com/mars-protocol/core-contracts)
[2](https://github.com/mars-protocol)
[3](https://docs.marsprotocol.io/smart-contracts/red-bank)
[4](https://forum.marsprotocol.io/t/an-overview-of-a-potential-mars-hub-architecture/559)
[5](https://cosmwasm.cosmos.network)
[6](https://cosmwasm.com)
[7](https://www.halborn.com/audits/mars-protocol/mars-protocol-core-contracts-updated-code-cosmwasm-smart-contract-security-assessment)
[8](https://www.linkedin.com/pulse/building-deploying-simple-cosmwasm-smart-contract-cosmos-rathore)
[9](https://defi-planet.com/2022/12/introducing-the-mars-protocol-an-open-source-protocol-on-cosmos/)
[10](https://cosmwasm.b9lab.com/01-introduction/03-integration.html)
[11](https://docs.mantrachain.io/developing-on-mantra-chain/cosmwasm-quick-start-guide/writing-and-deploying-cw20-contract)
[12](https://docs.cosmos.network/v0.46/CosmWasm/)
[13](https://blog.marsprotocol.io/blog/mars-protocol-code-review-breakdown-session-3-4-safety-fund-and-governance-modules)
[14](https://flagship.fyi/outposts/dapps/what-is-mars-protocol-hub-cosmos/)
[15](https://cosmwasm.com/build)
[16](https://cosmwasm.cosmos.network/tutorial)
[17](https://github.com/CosmWasm/cosmwasm)
[18](https://book.cosmwasm.com)
[19](https://crates.io/crates/mars-red-bank-types)
[20](https://docs.mantrachain.io/developing-on-mantra-chain/cosmwasm-quick-start-guide/cosmwasm-contracts)
