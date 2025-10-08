package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

// EndBlocker processes founder claim deadlines each block.
func (k Keeper) EndBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.processFounderDeadlines(sdkCtx)
}

func (k Keeper) processFounderDeadlines(ctx sdk.Context) error {
	now := uint64(ctx.BlockTime().Unix())

	var toProcess []struct {
		Deadline uint64
		Denom    string
	}

	k.iterateFounderDeadline(ctx, func(deadline uint64, denom string) bool {
		if deadline == 0 || deadline > now {
			return true
		}
		toProcess = append(toProcess, struct {
			Deadline uint64
			Denom    string
		}{Deadline: deadline, Denom: denom})
		return false
	})

	if len(toProcess) == 0 {
		return nil
	}

	params := k.GetParams(ctx)

	for _, item := range toProcess {
		token, found := k.getToken(ctx, item.Denom)
		if !found {
			k.removeFounderDeadline(ctx, item.Deadline, item.Denom)
			continue
		}

		if token.Distribution.FounderClaimed {
			k.removeFounderDeadline(ctx, item.Deadline, item.Denom)
			continue
		}

		founderReserved := token.Distribution.FounderReservedInt()
		if !founderReserved.IsZero() {
			if params.BondingCurveWallet == "" {
				return types.ErrParamAddresses
			}
			if err := k.mintToAddress(ctx, token.Creator, token.Denom, params.BondingCurveWallet, founderReserved); err != nil {
				return err
			}
			currentCurve := token.Distribution.BondingCurveSupplyInt()
			currentCurve = currentCurve.Add(founderReserved)
			token.Distribution.SetBondingCurveSupply(currentCurve)
		}

		token.Distribution.SetFounderReserved(osmomath.ZeroInt())
		token.Distribution.FounderClaimed = true

		k.setToken(ctx, token)
		k.removeFounderDeadline(ctx, item.Deadline, item.Denom)

		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeFounderClaim,
			sdk.NewAttribute(types.AttributeKeyDenom, token.Denom),
			sdk.NewAttribute(types.AttributeKeyFounderClaimed, "expired"),
		))
	}

	return nil
}
