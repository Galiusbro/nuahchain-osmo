package keeper

import (
	"strconv"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
)

type msgServer struct {
	Keeper
}

// Message server implementation (used by grpcMsgServer)

// ExchangeTokens handles token exchange transactions
func (k msgServer) ExchangeTokens(ctx sdk.Context, msg *types.MsgExchangeTokens) (*types.MsgExchangeTokensResponse, error) {

	// Get module parameters
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// Check if exchange is enabled
	if !params.Enabled {
		return nil, types.ErrExchangeDisabled
	}

	// Validate sender address
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	// Check if token is supported by USD Oracle
	if !k.usdOracleKeeper.IsTokenSupported(ctx, msg.TokenIn.Denom) {
		return nil, errors.Wrapf(types.ErrUnsupportedToken, "token %s is not supported", msg.TokenIn.Denom)
	}

	// Get current exchange rate with price deviation check
	exchangeRate, err := k.GetExchangeRate(ctx, msg.TokenIn.Denom)
	if err != nil {
		// Try to update exchange rate if not found (includes TWAP validation)
		if updateErr := k.UpdateExchangeRate(ctx, msg.TokenIn.Denom); updateErr != nil {
			return nil, errors.Wrapf(types.ErrExchangeRateNotFound, "failed to get exchange rate for %s: %s", msg.TokenIn.Denom, err)
		}
		exchangeRate, err = k.GetExchangeRate(ctx, msg.TokenIn.Denom)
		if err != nil {
			return nil, err
		}
	}

	// Additional real-time price deviation check
	tokenPrice, found := k.usdOracleKeeper.GetTokenPriceForExchange(ctx, msg.TokenIn.Denom)
	if found {
		// Validate current Oracle price against TWAP (using pool ID 1 as default)
		if err := k.ValidatePriceDeviation(ctx, msg.TokenIn.Denom, tokenPrice.Price, 1); err != nil {
			return nil, errors.Wrapf(err, "price validation failed for %s", msg.TokenIn.Denom)
		}
	}

	// Calculate USD value of input tokens
	usdValue := math.LegacyNewDecFromInt(msg.TokenIn.Amount).Mul(exchangeRate.Rate)

	// Validate exchange amount limits
	if usdValue.LT(params.MinExchangeAmountUsd) {
		return nil, errors.Wrapf(types.ErrExchangeAmountTooSmall, "exchange amount %s USD is below minimum %s USD", usdValue, params.MinExchangeAmountUsd)
	}

	if usdValue.GT(params.MaxExchangeAmountUsd) {
		return nil, errors.Wrapf(types.ErrExchangeAmountTooLarge, "exchange amount %s USD is above maximum %s USD", usdValue, params.MaxExchangeAmountUsd)
	}

	// Check daily limits
	today := k.GetTodayString(ctx)
	dailyLimit, err := k.GetDailyLimit(ctx, msg.Sender, today)
	if err != nil {
		// Create new daily limit if not found
		dailyLimit = types.DailyLimit{
			Address:            msg.Sender,
			TotalExchangedUsd:  math.LegacyZeroDec(),
			Date:               today,
		}
	}

	newDailyTotal := dailyLimit.TotalExchangedUsd.Add(usdValue)
	if newDailyTotal.GT(params.DailyLimitUsd) {
		return nil, errors.Wrapf(types.ErrDailyLimitExceeded, "daily limit exceeded: current %s + new %s > limit %s USD", dailyLimit.TotalExchangedUsd, usdValue, params.DailyLimitUsd)
	}

	// Calculate exchange fee
	feeAmount := usdValue.Mul(params.ExchangeFee)
	netUsdValue := usdValue.Sub(feeAmount)

	// Calculate N$ output (1:1 with USD after fees)
	nuahOut := netUsdValue.TruncateInt()

	// Check minimum output
	if nuahOut.LT(msg.MinNuahOut) {
		return nil, errors.Wrapf(types.ErrInvalidMinOutput, "output amount %s is less than minimum %s", nuahOut, msg.MinNuahOut)
	}

	// Check sender balance
	balance := k.bankKeeper.GetBalance(ctx, senderAddr, msg.TokenIn.Denom)
	if balance.Amount.LT(msg.TokenIn.Amount) {
		return nil, errors.Wrapf(types.ErrInsufficientBalance, "insufficient balance: have %s, need %s", balance.Amount, msg.TokenIn.Amount)
	}

	// Execute the exchange
	// 1. Send input tokens from user to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(msg.TokenIn)); err != nil {
		return nil, errors.Wrap(err, "failed to send tokens to module")
	}

	// 2. Mint N$ tokens
	nuahCoin := sdk.NewCoin("unuah", nuahOut) // Assuming N$ denom is "unuah"
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(nuahCoin)); err != nil {
		return nil, errors.Wrap(err, "failed to mint N$ tokens")
	}

	// 3. Send N$ tokens to user
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, sdk.NewCoins(nuahCoin)); err != nil {
		return nil, errors.Wrap(err, "failed to send N$ tokens to user")
	}

	// 4. Send input tokens to treasury (community pool)
	if len(params.TreasuryAddresses) > 0 {
		// For now, send to the first treasury address
		treasuryAddr, err := sdk.AccAddressFromBech32(params.TreasuryAddresses[0])
		if err != nil {
			return nil, errors.Wrapf(types.ErrInvalidTreasuryAddress, "invalid treasury address: %s", err)
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, treasuryAddr, sdk.NewCoins(msg.TokenIn)); err != nil {
			return nil, errors.Wrap(err, "failed to send tokens to treasury")
		}
	} else {
		// Send to community pool via distribution module
		if err := k.distributionKeeper.FundCommunityPool(ctx, sdk.NewCoins(msg.TokenIn), senderAddr); err != nil {
			return nil, errors.Wrap(err, "failed to send tokens to community pool")
		}
	}

	// 5. Update daily limit
	dailyLimit.TotalExchangedUsd = newDailyTotal
	if err := k.SetDailyLimit(ctx, dailyLimit); err != nil {
		return nil, errors.Wrap(err, "failed to update daily limit")
	}

	// Emit events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"exchange_tokens",
			sdk.NewAttribute("sender", msg.Sender),
			sdk.NewAttribute("token_in", msg.TokenIn.String()),
			sdk.NewAttribute("nuah_out", nuahOut.String()),
			sdk.NewAttribute("exchange_rate", exchangeRate.Rate.String()),
			sdk.NewAttribute("usd_value", usdValue.String()),
			sdk.NewAttribute("fee_amount", feeAmount.String()),
		),
	})

	return &types.MsgExchangeTokensResponse{
		NuahOut: nuahOut,
	}, nil
}

// UpdateParams handles parameter updates
func (k msgServer) UpdateParams(ctx sdk.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {

	if k.GetAuthority() != msg.Authority {
		return nil, errors.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"update_params",
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("enabled", strconv.FormatBool(msg.Params.Enabled)),
		),
	)

	return &types.MsgUpdateParamsResponse{}, nil
}