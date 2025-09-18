use cosmwasm_std::{coins, Addr, Empty, Uint128};
use cw_multi_test::{App, Contract, ContractWrapper, Executor};

use simple_counter::msg::{ExecuteMsg, GetCountResponse, InstantiateMsg, QueryMsg};

fn counter_contract() -> Box<dyn Contract<Empty>> {
    let contract = ContractWrapper::new(
        simple_counter::contract::execute,
        simple_counter::contract::instantiate,
        simple_counter::contract::query,
    );
    Box::new(contract)
}

#[test]
fn test_counter_integration() {
    let mut app = App::default();

    // Store the contract code
    let code_id = app.store_code(counter_contract());

    // Create an owner address
    let owner = Addr::unchecked("owner");

    // Instantiate the contract
    let contract_addr = app
        .instantiate_contract(
            code_id,
            owner.clone(),
            &InstantiateMsg { count: 0 },
            &[],
            "Counter",
            None,
        )
        .unwrap();

    // Query initial count
    let count_response: GetCountResponse = app
        .wrap()
        .query_wasm_smart(contract_addr.clone(), &QueryMsg::GetCount {})
        .unwrap();
    assert_eq!(count_response.count, 0);

    // Increment the counter
    app.execute_contract(
        owner.clone(),
        contract_addr.clone(),
        &ExecuteMsg::Increment {},
        &[],
    )
    .unwrap();

    // Query count after increment
    let count_response: GetCountResponse = app
        .wrap()
        .query_wasm_smart(contract_addr.clone(), &QueryMsg::GetCount {})
        .unwrap();
    assert_eq!(count_response.count, 1);

    // Reset the counter
    app.execute_contract(
        owner.clone(),
        contract_addr.clone(),
        &ExecuteMsg::Reset { count: 42 },
        &[],
    )
    .unwrap();

    // Query count after reset
    let count_response: GetCountResponse = app
        .wrap()
        .query_wasm_smart(contract_addr.clone(), &QueryMsg::GetCount {})
        .unwrap();
    assert_eq!(count_response.count, 42);

    // Try to reset from unauthorized user (should fail)
    let unauthorized = Addr::unchecked("unauthorized");
    let result = app.execute_contract(
        unauthorized,
        contract_addr.clone(),
        &ExecuteMsg::Reset { count: 100 },
        &[],
    );
    assert!(result.is_err());

    // Count should remain unchanged
    let count_response: GetCountResponse = app
        .wrap()
        .query_wasm_smart(contract_addr, &QueryMsg::GetCount {})
        .unwrap();
    assert_eq!(count_response.count, 42);
}