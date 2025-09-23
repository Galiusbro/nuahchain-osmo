Вот подробная инструкция по интеграции Umee CosmWasm на свой Cosmos SDK блокчейн:

***

### 1. Предварительные условия

- Ваша сеть должна поддерживать CosmWasm (модуль x/wasm должен быть уже интегрирован в SDK).[1][2][3]
- Установите Rust (nightly), cargo-make, Docker, Node.js для сборки и деплоя контрактов.

***

### 2. Сборка CosmWasm-контрактов Umee

- Клонируйте репозиторий Umee, перейдите в папку с CosmWasm-модулями (например, leverage, oracle, cw-umee-types).[4][5]
- Соберите .wasm файлы:

```bash
cargo install --force cargo-make
cargo make rust-optimizer
```

Готовые контракты появятся в каталоге `/artifacts`.

***

### 3. Загрузка (store) и деплой контракта

- Загрузите контракт в свою сеть с помощью CLI или REST:

```bash
wasmd tx wasm store ./artifacts/<contract>.wasm --from <your-key> --gas auto -y
```

Запишите выданный code_id.

- Проинстанциируйте контракт (настройте аргументы и admin):

```bash
wasmd tx wasm instantiate <code_id> '<init_msg>' --from <your-key> --label "umee-lending" --gas auto -y
```

Получите адрес контракта.

***

### 4. Настройка связки модулей и параметров

- Настройте взаимодействие контракта с вашим Bank Module.
- Проверьте доступность всех используемых токенов как collaterals.
- Привяжите или внедрите оракл-прод-устройки (вы можете подключить Cosmos Oracles либо кастомные через cw-umee-types).

***

### 5. Интеграция IBC, если потребуются кросс-чейн операции

- Настройте и протестируйте IBC-каналы между вашими цепями.
- Проверьте работу кросс-чейн депозита, borrow-repay, ликвидаций.[6]

***

### 6. Тестирование business-логики

- Запустите unit-тесты для supply, borrow, repay, liquidation через wasmcli или web-интерфейс.
- Проверьте безопасность и корректность на тестовой сети.

***

### 7. Управление и обновление

- Управляйте контрактом через ваш DAO, либо с помощью стандартных функций миграции CosmWasm.
- Обновляйте версии по мере выхода новых релизов Umee (Apache 2.0 — обновления свободны).

***

**Документация и ресурсы:**
- CosmWasm интеграция: https://cosmwasm.b9lab.com/01-introduction/03-integration.html[2]
- Контракты Umee: https://github.com/umee-network/umee[4]
- CosmWasm CLI: https://cosmwasm.com/build[7]

***

Эта инструкция охватывает интеграцию CosmWasm-контрактов Umee на свою сеть Cosmos SDK с возможностью развёртывания собственного бизнеса lending/borrowing под любыми активами без лицензионных ограничений.[3][5][1][2][6][4]

[1](https://cosmwasm.cosmos.network)
[2](https://cosmwasm.b9lab.com/01-introduction/03-integration.html)
[3](https://docs.cosmos.network/v0.46/CosmWasm/)
[4](https://github.com/umee-network/umee)
[5](https://crates.io/crates/cw-umee-types)
[6](https://hackmd.io/@robert-zaremba/S1wPFYaeT)
[7](https://cosmwasm.com)
[8](https://www.halborn.com/audits/umee/wasm-integration-cosmos-security-assessment)
[9](https://docs.mantrachain.io/developing-on-mantra-chain/cosmwasm-quick-start-guide/writing-and-deploying-cw20-contract)
[10](https://cosmwasm.cosmos.network/tutorial)
[11](https://github.com/CosmWasm/wasmd)
[12](https://docs.burnt.com/xion/developers/getting-started-advanced/your-first-contract/deploy-a-cosmwasm-smart-contract)
[13](https://cosmwasm.cosmos.network/core/installation)
[14](https://cosmwasm.com/build)
[15](https://www.range.org/blog/advanced-transaction-analysis-in-cosmos)
[16](https://docs.soarchain.com/smart-contracts/Deploy-a-Smart-Contract/)
[17](https://cosmwasm.b9lab.com/01-introduction/12-cross-contract.html)
[18](https://www.youtube.com/watch?v=VTjiC4wcd7k)
[19](https://docs.archway.io/developers/guides/my-first-dapp/deploy)
[20](https://book.cosmwasm.com/setting-up-env.html)
[21](https://book.cosmwasm.com/basics/building-contract.html)

