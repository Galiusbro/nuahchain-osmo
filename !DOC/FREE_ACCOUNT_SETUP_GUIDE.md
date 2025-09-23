# Free Account Setup Guide

This guide explains how to set up a blockchain node with free accounts that can send transactions without paying gas fees.

## What are Free Accounts?

Free accounts are special blockchain accounts that can send transactions without paying gas fees. This is useful for:
- New users who don't have tokens yet
- Testing and development
- Specific use cases where you want to subsidize transaction costs

## Prerequisites

- Built blockchain binary (`./build/nuahd`)
- Basic command line knowledge
- Understanding of blockchain concepts (accounts, transactions, genesis)

## Step-by-Step Setup

### 1. Clean Start

First, remove any existing blockchain data and start fresh:

```bash
# Remove existing data
rm -rf ~/.nuahd

# Initialize the node with your chain ID
./build/nuahd init validator --chain-id nuahchain-1
```

### 2. Create Validator Key

Create a validator key that will be used to secure the network:

```bash
# Create validator key (you'll be prompted for a passphrase)
./build/nuahd keys add validator

# Get the validator address (save this for later)
./build/nuahd keys show validator -a
```

### 3. Update Genesis Configuration

Modify the genesis file to use your custom token denomination:

```bash
# Replace 'stake' with 'unuah' in genesis.json
sed -i '' 's/"stake"/"unuah"/g' ~/.nuahd/config/genesis.json
```

### 4. Add Genesis Accounts

Add accounts with initial balances to the genesis file:

```bash
# Add validator account with 10,000,000 NUAH (10,000,000,000,000 unuah)
./build/nuahd add-genesis-account validator 10000000000000unuah

# Add test user account with 1,000,000 unuah
./build/nuahd add-genesis-account testuser 1000000unuah
```

### 5. Generate Genesis Transaction

Create the initial validator transaction:

```bash
# Generate gentx for validator with 5,000,000 NUAH stake
./build/nuahd gentx validator 5000000000000unuah --chain-id nuahchain-1

# Collect all genesis transactions
./build/nuahd collect-gentxs
```

### 6. Start the Node

Launch your blockchain node:

```bash
# Start the node (this will run continuously)
./build/nuahd start
```

Keep this terminal open. The node will start producing blocks.

### 7. Verify Free Account Status

In a new terminal, check if your test account is configured as a free account:

```bash
# Check if testuser is a free account
./build/nuahd query freeaccount is-free-account [TESTUSER_ADDRESS]
```

Replace `[TESTUSER_ADDRESS]` with the actual address. You should see:
```json
{
  "is_free": true
}
```

### 8. Test Free Transactions

Now test that the free account can send transactions without fees:

```bash
# Send transaction without specifying fees
./build/nuahd tx bank send testuser [RECIPIENT_ADDRESS] 1000unuah \
  --chain-id nuahchain-1 \
  --gas auto \
  --gas-adjustment 1.5 \
  -y
```

### 9. Verify Transaction Success

Check that the transaction was successful and no fees were deducted:

```bash
# Check sender balance
./build/nuahd query bank balances [TESTUSER_ADDRESS]

# Check recipient balance
./build/nuahd query bank balances [RECIPIENT_ADDRESS]
```

## Expected Results

After following this guide:

1. **Free Account Status**: The testuser account should show `"is_free": true`
2. **No Fee Deduction**: When sending transactions, only the sent amount is deducted from the balance, not additional gas fees
3. **Successful Transactions**: All transactions should complete with exit code 0

## Example Test Results

```
Initial testuser balance: 1,000,000 unuah
After sending 1,000 unuah: 999,000 unuah (only sent amount deducted)
After sending 5,000 unuah: 994,000 unuah (still no gas fees)
```

## Troubleshooting

### Genesis Validation Fails
If `./build/nuahd validate-genesis` fails, you can still proceed. The node may start successfully despite validation warnings.

### Account Not Free
If the account doesn't show as free:
1. Check that the account was added to genesis correctly
2. Verify the freeaccount module is enabled in your build
3. Ensure the account address matches exactly

### Transaction Fails
If transactions fail:
1. Make sure the node is running and synced
2. Check account balances are sufficient
3. Verify the chain-id matches your genesis configuration

## Technical Details

### How It Works

1. **Genesis Configuration**: Free accounts are configured in the genesis file under the `freeaccount` module
2. **Ante Handler**: The blockchain's ante handler checks if an account is free before deducting gas fees
3. **Module Integration**: The freeaccount module integrates with the bank module to enable fee-free transactions

### Key Files

- `~/.nuahd/config/genesis.json`: Contains the initial blockchain state including free accounts
- `~/.nuahd/config/app.toml`: Node configuration (gas prices, mempool settings)
- `~/.nuahd/config/config.toml`: Consensus and networking configuration

## Security Considerations

- Free accounts should be used carefully in production
- Consider rate limiting or other protections against spam
- Monitor free account usage to prevent abuse
- Regularly review which accounts have free status

## Next Steps

After setting up free accounts, you might want to:

1. Configure additional node settings in `app.toml`
2. Set up monitoring and logging
3. Add more validator nodes for decentralization
4. Implement additional security measures
5. Create a user interface for managing free accounts

This completes the free account setup. Your blockchain now supports accounts that can transact without paying gas fees!