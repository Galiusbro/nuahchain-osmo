package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

// InitGenesis initializes state from genesis data.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	for _, token := range genState.Tokens {
		k.setToken(ctx, token)
		if !token.Distribution.FounderClaimed && token.Distribution.FounderClaimDeadline > 0 {
			k.queueFounderDeadline(ctx, token.Distribution.FounderClaimDeadline, token.Denom)
		}
	}
}

// ExportGenesis exports module state to genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	tokens := make([]types.Token, 0)

	k.IterateTokens(ctx, func(token types.Token) bool {
		tokens = append(tokens, token)
		return false
	})

	return &types.GenesisState{
		Params: params,
		Tokens: tokens,
	}
}
