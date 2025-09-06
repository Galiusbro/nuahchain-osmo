package softpeg

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	softpegkeeper "github.com/osmosis-labs/osmosis/v30/x/softpeg/keeper"
	softpegtypes "github.com/osmosis-labs/osmosis/v30/x/softpeg/types"
)

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx sdk.Context, k softpegkeeper.Keeper) {
	// Check if monitoring is enabled
	if !k.MonitoringEnabled(ctx) {
		return
	}

	// Get the last update time from store or use genesis time
	lastUpdateKey := []byte("last_price_update")
	store := ctx.KVStore(k.GetStoreKey())
	lastUpdateBytes := store.Get(lastUpdateKey)

	var lastUpdate time.Time
	if lastUpdateBytes != nil {
		lastUpdate = time.Unix(int64(sdk.BigEndianToUint64(lastUpdateBytes)), 0)
	} else {
		// First time running, use genesis time
		lastUpdate = ctx.BlockTime().Add(-k.UpdateInterval(ctx))
	}

	// Check if it's time to update
	if ctx.BlockTime().Sub(lastUpdate) >= k.UpdateInterval(ctx) {
		// Update price data
		if err := k.UpdatePriceData(ctx); err != nil {
			k.Logger(ctx).Error("Failed to update price data", "error", err)
		} else {
			// Store the update time
			store.Set(lastUpdateKey, sdk.Uint64ToBigEndian(uint64(ctx.BlockTime().Unix())))
			k.Logger(ctx).Info("Price data updated successfully")
		}

		// Check for price deviations and create alerts if necessary
		if deviation, deviationAmount, err := k.CheckPriceDeviation(ctx); err == nil {
			if deviation {
				k.Logger(ctx).Warn("Price deviation detected",
					"deviation", deviationAmount.String(),
					"threshold", k.DeviationThreshold(ctx).String())

				// Create alert if deviation exceeds alert threshold
				if deviationAmount.GT(k.AlertThreshold(ctx)) {
					k.CreatePriceAlert(ctx, deviationAmount)
				}
			}
		} else {
			k.Logger(ctx).Error("Failed to check price deviation", "error", err)
		}
	}

	// Clean up old data periodically (every 100 blocks)
	if ctx.BlockHeight()%100 == 0 {
		k.DeleteOldPriceData(ctx)
		k.Logger(ctx).Debug("Old price data cleaned up")
	}
}

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k softpegkeeper.Keeper) {
	// Update community metrics if needed
	if ctx.BlockHeight()%10 == 0 { // Update every 10 blocks
		if err := k.UpdateCommunityMetrics(ctx); err != nil {
			k.Logger(ctx).Error("Failed to update community metrics", "error", err)
		} else {
			k.Logger(ctx).Debug("Community metrics updated")
		}
	}

	// Emit events for monitoring
	if latestPrice, found := k.GetLatestPriceData(ctx); found {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				softpegtypes.EventTypePriceUpdate,
				sdk.NewAttribute(softpegtypes.AttributeKeyPrice, latestPrice.Price.String()),
				sdk.NewAttribute(softpegtypes.AttributeKeyConfidence, latestPrice.Confidence.String()),
				sdk.NewAttribute(softpegtypes.AttributeKeyTimestamp, fmt.Sprintf("%d", latestPrice.Timestamp)),
			),
		)
	}
}
