package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"

	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		// Keepers
		accountKeeper      types.AccountKeeper
		bankKeeper         types.BankKeeper
		usdOracleKeeper    types.USDOracleKeeper
		twapKeeper         types.TWAPKeeper
		distributionKeeper types.DistributionKeeper

		// Collections
		ParamsStore        collections.Item[types.Params]
		ExchangeRatesStore collections.Map[string, types.ExchangeRate]
		DailyLimits        collections.Map[collections.Pair[string, string], types.DailyLimit] // address + date -> limit
		schema             collections.Schema
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	usdOracleKeeper types.USDOracleKeeper,
	twapKeeper types.TWAPKeeper,
	distributionKeeper types.DistributionKeeper,
) *Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:                cdc,
		storeService:       storeService,
		logger:             logger,
		authority:          authority,
		accountKeeper:      accountKeeper,
		bankKeeper:         bankKeeper,
		usdOracleKeeper:    usdOracleKeeper,
		twapKeeper:         twapKeeper,
		distributionKeeper: distributionKeeper,
		ParamsStore: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		ExchangeRatesStore: collections.NewMap(
			sb,
			types.ExchangeRateKeyPrefix,
			"exchange_rates",
			collections.StringKey,
			codec.CollValue[types.ExchangeRate](cdc),
		),
		DailyLimits: collections.NewMap(
			sb,
			types.DailyLimitKeyPrefix,
			"daily_limits",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[types.DailyLimit](cdc),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.schema = schema

	return &k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	params, err := k.ParamsStore.Get(ctx)
	if err != nil {
		return types.Params{}, err
	}
	return params, nil
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return k.ParamsStore.Set(ctx, params)
}

// GetExchangeRate returns the exchange rate for a given denom
func (k Keeper) GetExchangeRate(ctx sdk.Context, denom string) (types.ExchangeRate, error) {
	return k.ExchangeRatesStore.Get(ctx, denom)
}

// SetExchangeRate sets the exchange rate for a given denom
func (k Keeper) SetExchangeRate(ctx sdk.Context, rate types.ExchangeRate) error {
	return k.ExchangeRatesStore.Set(ctx, rate.Denom, rate)
}

// GetAllExchangeRates returns all exchange rates
func (k Keeper) GetAllExchangeRates(ctx sdk.Context) ([]types.ExchangeRate, error) {
	var rates []types.ExchangeRate
	err := k.ExchangeRatesStore.Walk(ctx, nil, func(key string, value types.ExchangeRate) (bool, error) {
		rates = append(rates, value)
		return false, nil
	})
	return rates, err
}

// GetDailyLimit returns the daily limit for a given address and date
func (k Keeper) GetDailyLimit(ctx sdk.Context, address, date string) (types.DailyLimit, error) {
	key := collections.Join(address, date)
	return k.DailyLimits.Get(ctx, key)
}

// SetDailyLimit sets the daily limit for a given address and date
func (k Keeper) SetDailyLimit(ctx sdk.Context, limit types.DailyLimit) error {
	key := collections.Join(limit.Address, limit.Date)
	return k.DailyLimits.Set(ctx, key, limit)
}

// GetTodayString returns today's date in YYYY-MM-DD format
func (k Keeper) GetTodayString(ctx sdk.Context) string {
	return ctx.BlockTime().Format("2006-01-02")
}

// GetTWAPPrice gets TWAP price for a token from the TWAP module
func (k Keeper) GetTWAPPrice(ctx sdk.Context, denom string, poolId uint64) (osmomath.Dec, error) {
	// Get TWAP for the last 15 minutes
	endTime := ctx.BlockTime()
	startTime := endTime.Add(-15 * time.Minute)

	// Use USDC as quote asset for TWAP calculation
	quoteAsset := "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858" // USDC

	twapPrice, err := k.twapKeeper.GetArithmeticTwap(ctx, poolId, denom, quoteAsset, startTime, endTime)
	if err != nil {
		return osmomath.ZeroDec(), err
	}

	return twapPrice, nil
}

// ValidatePriceDeviation checks if Oracle price deviates too much from TWAP price
func (k Keeper) ValidatePriceDeviation(ctx sdk.Context, denom string, oraclePrice math.LegacyDec, poolId uint64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// Get TWAP price
	twapPriceOsmo, err := k.GetTWAPPrice(ctx, denom, poolId)
	if err != nil {
		// If TWAP is not available, log warning but allow Oracle price
		k.Logger().Warn("TWAP price not available, using Oracle price only", "denom", denom, "error", err)
		return nil
	}

	// Convert osmomath.Dec to math.LegacyDec for comparison
	twapPrice := math.LegacyNewDecFromBigInt(twapPriceOsmo.BigInt())

	// Calculate deviation percentage
	var deviation math.LegacyDec
	if oraclePrice.GT(twapPrice) {
		deviation = oraclePrice.Sub(twapPrice).Quo(twapPrice)
	} else {
		deviation = twapPrice.Sub(oraclePrice).Quo(twapPrice)
	}

	// Check if deviation exceeds threshold
	if deviation.GT(params.PriceDeviationThreshold) {
		return types.ErrPriceDeviationTooHigh
	}

	return nil
}

// UpdateExchangeRate updates the exchange rate for a token using Oracle and TWAP data
func (k Keeper) UpdateExchangeRate(ctx sdk.Context, denom string) error {
	// Get Oracle price
	tokenPrice, found := k.usdOracleKeeper.GetTokenPriceForExchange(ctx, denom)
	if !found {
		return types.ErrExchangeRateNotFound
	}

	// Validate price deviation (using pool ID 1 as default, should be configurable)
	if err := k.ValidatePriceDeviation(ctx, denom, tokenPrice.Price, 1); err != nil {
		return err
	}

	rate := types.ExchangeRate{
		Denom:       denom,
		Rate:        tokenPrice.Price,
		LastUpdated: time.Now(),
	}

	return k.SetExchangeRate(ctx, rate)
}
