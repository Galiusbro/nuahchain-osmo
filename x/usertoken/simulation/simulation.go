package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/osmosis-labs/osmosis/v30/x/usertoken/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateUserToken = "op_weight_msg_create_user_token"
	OpWeightMsgBuyTokens       = "op_weight_msg_buy_tokens"
	OpWeightMsgSellTokens      = "op_weight_msg_sell_tokens"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper,
	bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	return simulation.WeightedOperations{}
}

// SimulateMsgCreateUserToken generates a MsgCreateUserToken with random values
func SimulateMsgCreateUserToken(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateUserToken, "simulation not implemented"), nil, nil
	}
}

// SimulateMsgBuyTokens generates a MsgBuyTokens with random values
func SimulateMsgBuyTokens(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBuyTokens, "simulation not implemented"), nil, nil
	}
}

// SimulateMsgSellTokens generates a MsgSellTokens with random values
func SimulateMsgSellTokens(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSellTokens, "simulation not implemented"), nil, nil
	}
}
