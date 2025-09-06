package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	softpegTypes "github.com/osmosis-labs/osmosis/v30/x/softpeg/types"
)

type (
	// Keeper of the softpeg store
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		// External keepers
		bankKeeper   softpegTypes.BankKeeper
		poolKeeper   softpegTypes.PoolKeeper
		gammKeeper   softpegTypes.GAMMKeeper
		epochsKeeper softpegTypes.EpochsKeeper
		govKeeper    softpegTypes.GovKeeper
	}
)

// NewKeeper creates a new softpeg Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper softpegTypes.BankKeeper,
	poolKeeper softpegTypes.PoolKeeper,
	gammKeeper softpegTypes.GAMMKeeper,
	epochsKeeper softpegTypes.EpochsKeeper,
	govKeeper softpegTypes.GovKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(softpegTypes.ParamKeyTable())
	}

	return &Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		memKey:       memKey,
		paramstore:   ps,
		bankKeeper:   bankKeeper,
		poolKeeper:   poolKeeper,
		gammKeeper:   gammKeeper,
		epochsKeeper: epochsKeeper,
		govKeeper:    govKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", softpegTypes.ModuleName))
}

// GetParams get all parameters as softpegTypes.Params
func (k Keeper) GetParams(ctx sdk.Context) softpegTypes.Params {
	return softpegTypes.NewParams(
		k.TargetPrice(ctx),
		k.DeviationThreshold(ctx),
		k.AlertThreshold(ctx),
		k.MonitoringEnabled(ctx),
		k.UpdateInterval(ctx),
		k.MinLiquidity(ctx),
		k.MaxDeviation(ctx),
		k.TrustDecayRate(ctx),
		k.MinConfidence(ctx),
		k.MaxAlertsPerHour(ctx),
		k.FeedbackCooldown(ctx),
		k.PriceDataRetention(ctx),
		k.CommunityWeight(ctx),
		k.LiquidityWeight(ctx),
		k.VolumeWeight(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params softpegTypes.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// TargetPrice returns the TargetPrice param
func (k Keeper) TargetPrice(ctx sdk.Context) (res osmomath.Dec) {
	k.paramstore.Get(ctx, softpegTypes.KeyTargetPrice, &res)
	return
}

// DeviationThreshold returns the DeviationThreshold param
func (k Keeper) DeviationThreshold(ctx sdk.Context) (res osmomath.Dec) {
	k.paramstore.Get(ctx, softpegTypes.KeyDeviationThreshold, &res)
	return
}

// AlertThreshold returns the AlertThreshold param
func (k Keeper) AlertThreshold(ctx sdk.Context) (res osmomath.Dec) {
	k.paramstore.Get(ctx, softpegTypes.KeyAlertThreshold, &res)
	return
}

// MonitoringEnabled returns the MonitoringEnabled param
func (k Keeper) MonitoringEnabled(ctx sdk.Context) (res bool) {
	k.paramstore.Get(ctx, softpegTypes.KeyMonitoringEnabled, &res)
	return
}

// UpdateInterval returns the UpdateInterval param
func (k Keeper) UpdateInterval(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, softpegTypes.KeyUpdateInterval, &res)
	return
}

// MinLiquidity returns the MinLiquidity param
func (k Keeper) MinLiquidity(ctx sdk.Context) (res osmomath.Int) {
	k.paramstore.Get(ctx, softpegTypes.KeyMinLiquidity, &res)
	return
}

// MaxDeviation returns the MaxDeviation param
func (k Keeper) MaxDeviation(ctx sdk.Context) (res osmomath.Dec) {
	k.paramstore.Get(ctx, softpegTypes.KeyMaxDeviation, &res)
	return
}

// TrustDecayRate returns the TrustDecayRate param
func (k Keeper) TrustDecayRate(ctx sdk.Context) (res osmomath.Dec) {
	k.paramstore.Get(ctx, softpegTypes.KeyTrustDecayRate, &res)
	return
}

// MinConfidence returns the MinConfidence param
func (k Keeper) MinConfidence(ctx sdk.Context) (res osmomath.Dec) {
	k.paramstore.Get(ctx, softpegTypes.KeyMinConfidence, &res)
	return
}

// MaxAlertsPerHour returns the MaxAlertsPerHour param
func (k Keeper) MaxAlertsPerHour(ctx sdk.Context) (res int64) {
	k.paramstore.Get(ctx, softpegTypes.KeyMaxAlertsPerHour, &res)
	return
}

// FeedbackCooldown returns the FeedbackCooldown param
func (k Keeper) FeedbackCooldown(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, softpegTypes.KeyFeedbackCooldown, &res)
	return
}

// PriceDataRetention returns the PriceDataRetention param
func (k Keeper) PriceDataRetention(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, softpegTypes.KeyPriceDataRetention, &res)
	return
}

// CommunityWeight returns the CommunityWeight param
func (k Keeper) CommunityWeight(ctx sdk.Context) (res osmomath.Dec) {
	k.paramstore.Get(ctx, softpegTypes.KeyCommunityWeight, &res)
	return
}

// LiquidityWeight returns the LiquidityWeight param
func (k Keeper) LiquidityWeight(ctx sdk.Context) (res osmomath.Dec) {
	k.paramstore.Get(ctx, softpegTypes.KeyLiquidityWeight, &res)
	return
}

// VolumeWeight returns the VolumeWeight param
func (k Keeper) VolumeWeight(ctx sdk.Context) (res osmomath.Dec) {
	k.paramstore.Get(ctx, softpegTypes.KeyVolumeWeight, &res)
	return
}

// Price Data Management

// SetPriceData stores price data in the store
func (k Keeper) SetPriceData(ctx sdk.Context, priceData softpegTypes.PriceData) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&priceData)
	store.Set(softpegTypes.PriceDataKey(priceData.Timestamp), b)
}

// GetPriceData retrieves price data by timestamp
func (k Keeper) GetPriceData(ctx sdk.Context, timestamp int64) (softpegTypes.PriceData, bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(softpegTypes.PriceDataKey(timestamp))
	if b == nil {
		return softpegTypes.PriceData{}, false
	}

	var priceData softpegTypes.PriceData
	k.cdc.MustUnmarshal(b, &priceData)
	return priceData, true
}

// GetLatestPriceData retrieves the most recent price data
func (k Keeper) GetLatestPriceData(ctx sdk.Context) (softpegTypes.PriceData, bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStoreReversePrefixIterator(store, softpegTypes.PriceDataPrefix)
	defer iterator.Close()

	if !iterator.Valid() {
		return softpegTypes.PriceData{}, false
	}

	var priceData softpegTypes.PriceData
	k.cdc.MustUnmarshal(iterator.Value(), &priceData)
	return priceData, true
}

// GetAllPriceData retrieves all price data
func (k Keeper) GetAllPriceData(ctx sdk.Context) []softpegTypes.PriceData {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, softpegTypes.PriceDataPrefix)
	defer iterator.Close()

	var priceDataList []softpegTypes.PriceData
	for ; iterator.Valid(); iterator.Next() {
		var priceData softpegTypes.PriceData
		k.cdc.MustUnmarshal(iterator.Value(), &priceData)
		priceDataList = append(priceDataList, priceData)
	}

	return priceDataList
}

// DeleteOldPriceData removes price data older than retention period
func (k Keeper) DeleteOldPriceData(ctx sdk.Context) {
	retentionPeriod := k.PriceDataRetention(ctx)
	cutoffTime := ctx.BlockTime().Add(-retentionPeriod).Unix()

	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, softpegTypes.PriceDataPrefix)
	defer iterator.Close()

	var keysToDelete [][]byte
	for ; iterator.Valid(); iterator.Next() {
		var priceData softpegTypes.PriceData
		k.cdc.MustUnmarshal(iterator.Value(), &priceData)
		if priceData.Timestamp < cutoffTime {
			keysToDelete = append(keysToDelete, iterator.Key())
		}
	}

	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// Community Metrics Management

// SetCommunityMetrics stores community metrics in the store
func (k Keeper) SetCommunityMetrics(ctx sdk.Context, metrics softpegTypes.CommunityMetrics) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&metrics)
	store.Set(softpegTypes.CommunityMetricsKey(metrics.Timestamp), b)
}

// GetCommunityMetrics retrieves community metrics by timestamp
func (k Keeper) GetCommunityMetrics(ctx sdk.Context, timestamp int64) (softpegTypes.CommunityMetrics, bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(softpegTypes.CommunityMetricsKey(timestamp))
	if b == nil {
		return softpegTypes.CommunityMetrics{}, false
	}

	var metrics softpegTypes.CommunityMetrics
	k.cdc.MustUnmarshal(b, &metrics)
	return metrics, true
}

// GetLatestCommunityMetrics retrieves the most recent community metrics
func (k Keeper) GetLatestCommunityMetrics(ctx sdk.Context) (softpegTypes.CommunityMetrics, bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStoreReversePrefixIterator(store, softpegTypes.CommunityMetricsPrefix)
	defer iterator.Close()

	if !iterator.Valid() {
		return softpegTypes.CommunityMetrics{}, false
	}

	var metrics softpegTypes.CommunityMetrics
	k.cdc.MustUnmarshal(iterator.Value(), &metrics)
	return metrics, true
}

// Peg Configuration Management

// SetPegConfig stores peg configuration in the store
func (k Keeper) SetPegConfig(ctx sdk.Context, config softpegTypes.PegConfig) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&config)
	store.Set(softpegTypes.PegConfigKey(), b)
}

// GetPegConfig retrieves peg configuration
func (k Keeper) GetPegConfig(ctx sdk.Context) (softpegTypes.PegConfig, bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(softpegTypes.PegConfigKey())
	if b == nil {
		return softpegTypes.PegConfig{}, false
	}

	var config softpegTypes.PegConfig
	k.cdc.MustUnmarshal(b, &config)
	return config, true
}

// Price Monitoring and Analysis

// CalculateCurrentPrice calculates the current NUAH price based on pool data
func (k Keeper) CalculateCurrentPrice(ctx sdk.Context) (osmomath.Dec, error) {
	// Get NUAH pools from GAMM keeper
	pools := k.gammKeeper.GetPools(ctx)
	var totalLiquidity osmomath.Dec
	var weightedPrice osmomath.Dec

	for _, pool := range pools {
		// Check if pool contains NUAH
		if k.poolContainsNUAH(pool) {
			price, liquidity, err := k.getPoolPrice(ctx, pool)
			if err != nil {
				continue
			}
			weightedPrice = weightedPrice.Add(price.Mul(liquidity))
			totalLiquidity = totalLiquidity.Add(liquidity)
		}
	}

	if totalLiquidity.IsZero() {
		return osmomath.ZeroDec(), fmt.Errorf("no liquidity found for NUAH")
	}

	return weightedPrice.Quo(totalLiquidity), nil
}

// CheckPriceDeviation checks if current price deviates from target
func (k Keeper) CheckPriceDeviation(ctx sdk.Context) (bool, osmomath.Dec, error) {
	currentPrice, err := k.CalculateCurrentPrice(ctx)
	if err != nil {
		return false, osmomath.ZeroDec(), err
	}

	targetPrice := k.TargetPrice(ctx)
	deviation := currentPrice.Sub(targetPrice).Quo(targetPrice).Abs()
	threshold := k.DeviationThreshold(ctx)

	return deviation.GT(threshold), deviation, nil
}

// UpdatePriceData updates price data with current market information
func (k Keeper) UpdatePriceData(ctx sdk.Context) error {
	currentPrice, err := k.CalculateCurrentPrice(ctx)
	if err != nil {
		return err
	}

	// Calculate confidence based on liquidity and volume
	confidence := k.calculatePriceConfidence(ctx)

	priceData := softpegTypes.PriceData{
		Timestamp:  ctx.BlockTime().Unix(),
		Price:      currentPrice,
		Confidence: confidence,
		Source:     "pool_aggregation",
	}

	k.SetPriceData(ctx, priceData)
	return nil
}

// Helper functions

func (k Keeper) poolContainsNUAH(pool interface{}) bool {
	// Implementation depends on pool interface
	// This is a placeholder - actual implementation would check pool assets
	return true
}

func (k Keeper) getPoolPrice(ctx sdk.Context, pool interface{}) (osmomath.Dec, osmomath.Dec, error) {
	// Implementation depends on pool interface
	// This is a placeholder - actual implementation would calculate price and liquidity
	return osmomath.OneDec(), osmomath.NewDec(1000), nil
}

func (k Keeper) calculatePriceConfidence(ctx sdk.Context) osmomath.Dec {
	// Calculate confidence based on liquidity, volume, and other factors
	// This is a simplified implementation
	return osmomath.MustNewDecFromStr("0.85")
}

// GetStoreKey returns the store key for the softpeg module
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// CreatePriceAlert creates a price alert event
func (k Keeper) CreatePriceAlert(ctx sdk.Context, deviation osmomath.Dec) {
	var severity string
	if deviation.GT(k.MaxDeviation(ctx)) {
		severity = "critical"
	} else if deviation.GT(k.AlertThreshold(ctx)) {
		severity = "warning"
	} else {
		severity = "info"
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			softpegTypes.EventTypePriceAlert,
			sdk.NewAttribute(softpegTypes.AttributeKeyDeviation, deviation.String()),
			sdk.NewAttribute(softpegTypes.AttributeKeySeverity, severity),
			sdk.NewAttribute(softpegTypes.AttributeKeyAlertType, "price_deviation"),
			sdk.NewAttribute(softpegTypes.AttributeKeyTimestamp, osmomath.NewInt(ctx.BlockTime().Unix()).String()),
		),
	)
}

// UpdateCommunityMetrics updates community metrics
func (k Keeper) UpdateCommunityMetrics(ctx sdk.Context) error {
	// This is a placeholder implementation
	// In a real implementation, this would gather data from governance, social media APIs, etc.
	metrics := softpegTypes.CommunityMetrics{
		Timestamp:          ctx.BlockTime().Unix(),
		SentimentScore:     osmomath.MustNewDecFromStr("0.75"),
		TrustScore:         osmomath.MustNewDecFromStr("0.80"),
		ParticipationRate:  osmomath.MustNewDecFromStr("0.65"),
		GovernanceActivity: 10,
		SocialMentions:     100,
		CommunitySize:      1000,
	}

	k.SetCommunityMetrics(ctx, metrics)
	return nil
}
