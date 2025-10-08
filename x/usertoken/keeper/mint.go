package keeper

import (
	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v30/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v30/x/tokenfactory/types"
	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

func (k Keeper) mintToAddress(ctx sdk.Context, admin string, denom string, recipient string, amount osmomath.Int) error {
	if amount.IsZero() {
		return nil
	}

	if k.tokenFactoryKeeper == nil {
		return types.ErrTokenFactory.Wrap("token factory keeper not configured")
	}

	msgServer := tokenfactorykeeper.NewMsgServerImpl(*k.tokenFactoryKeeper)
	coin := sdk.NewCoin(denom, amount)
	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), tokenfactorytypes.NewMsgMintTo(admin, coin, recipient))
	if err != nil {
		return types.ErrTokenFactory.Wrap(err.Error())
	}
	return nil
}
