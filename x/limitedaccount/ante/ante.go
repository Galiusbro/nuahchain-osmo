package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	limitedaccountkeeper "github.com/osmosis-labs/osmosis/v30/x/limitedaccount/keeper"
	limitedaccounttypes "github.com/osmosis-labs/osmosis/v30/x/limitedaccount/types"
)

// LimitedAccountDecorator checks transaction limits for limited accounts
type LimitedAccountDecorator struct {
	limitedAccountKeeper *limitedaccountkeeper.Keeper
	cdc                  codec.Codec
}

// NewLimitedAccountDecorator creates a new LimitedAccountDecorator
func NewLimitedAccountDecorator(limitedAccountKeeper *limitedaccountkeeper.Keeper, cdc codec.Codec) LimitedAccountDecorator {
	return LimitedAccountDecorator{
		limitedAccountKeeper: limitedAccountKeeper,
		cdc:                  cdc,
	}
}

// AnteHandle checks if the transaction sender is a limited account and enforces daily limits
func (lad LimitedAccountDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Skip checks during simulation
	if simulate {
		return next(ctx, tx, simulate)
	}

	// Get all messages from the transaction
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return next(ctx, tx, simulate)
	}

	// Check each message for limited account signers
	for _, msg := range msgs {
		signers, _, err := lad.cdc.GetMsgV1Signers(msg)
		if err != nil {
			return ctx, err
		}
		
		for _, signer := range signers {
			address := sdk.AccAddress(signer).String()
			
			// Check if this is a limited account
			if lad.limitedAccountKeeper.IsLimitedAccount(ctx, address) {
				// Check if the account can transact
				if !lad.limitedAccountKeeper.CanTransact(ctx, address) {
					return ctx, limitedaccounttypes.ErrDailyLimitExceeded
				}
				
				// Process the transaction (increment counter)
				if err := lad.limitedAccountKeeper.ProcessTransaction(ctx, address); err != nil {
					return ctx, err
				}
			}
		}
	}

	return next(ctx, tx, simulate)
}