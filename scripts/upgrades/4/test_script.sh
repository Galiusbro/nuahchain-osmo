# Download a genesis.json for testing. The node that you this on will be your "validator"
# It should be on version v4.x

nuahd init --chain-id=testing testing --home=$HOME/.nuahd
nuahd keys add validator --keyring-backend=test --home=$HOME/.nuahd
nuahd add-genesis-account $(nuahd keys show validator -a --keyring-backend=test --home=$HOME/.nuahd) 1000000000uosmo,1000000000valtoken --home=$HOME/.nuahd
sed -i -e "s/stake/uosmo/g" $HOME/.nuahd/config/genesis.json
nuahd gentx validator 500000000uosmo --commission-rate="0.0" --keyring-backend=test --home=$HOME/.nuahd --chain-id=testing
nuahd collect-gentxs --home=$HOME/.nuahd

cat $HOME/.nuahd/config/genesis.json | jq '.initial_height="711800"' > $HOME/.nuahd/config/tmp_genesis.json && mv $HOME/.nuahd/config/tmp_genesis.json $HOME/.nuahd/config/genesis.json
cat $HOME/.nuahd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"]["denom"]="valtoken"' > $HOME/.nuahd/config/tmp_genesis.json && mv $HOME/.nuahd/config/tmp_genesis.json $HOME/.nuahd/config/genesis.json
cat $HOME/.nuahd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"]["amount"]="100"' > $HOME/.nuahd/config/tmp_genesis.json && mv $HOME/.nuahd/config/tmp_genesis.json $HOME/.nuahd/config/genesis.json
cat $HOME/.nuahd/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="120s"' > $HOME/.nuahd/config/tmp_genesis.json && mv $HOME/.nuahd/config/tmp_genesis.json $HOME/.nuahd/config/genesis.json
cat $HOME/.nuahd/config/genesis.json | jq '.app_state["staking"]["params"]["min_commission_rate"]="0.050000000000000000"' > $HOME/.nuahd/config/tmp_genesis.json && mv $HOME/.nuahd/config/tmp_genesis.json $HOME/.nuahd/config/genesis.json

# Now setup a second full node, and peer it with this v3.0.0-rc0 node.

# start the chain on both machines
nuahd start
# Create proposals

nuahd tx gov submit-proposal --title="existing passing prop" --description="passing prop"  --from=validator --deposit=1000valtoken --chain-id=testing --keyring-backend=test --broadcast-mode=block  --type="Text"
nuahd tx gov vote 1 yes --from=validator --keyring-backend=test --chain-id=testing --yes
nuahd tx gov submit-proposal --title="prop with enough osmo deposit" --description="prop w/ enough deposit"  --from=validator --deposit=500000000uosmo --chain-id=testing --keyring-backend=test --broadcast-mode=block  --type="Text"
# Check that we have proposal 1 passed, and proposal 2 in deposit period
nuahd q gov proposals
# CHeck that validator commission is under min_commission_rate
nuahd q staking validators
# Wait for upgrade block.
# Upgrade happened
# your full node should have crashed with consensus failure

# Now we test post-upgrade behavior is as intended

# Everything in deposit stayed in deposit
nuahd q gov proposals
# Check that commissions was bumped to min_commission_rate
nuahd q staking validators
# pushes 2 into voting period
nuahd tx gov deposit 2 1valtoken --from=validator --keyring-backend=test --chain-id=testing --yes