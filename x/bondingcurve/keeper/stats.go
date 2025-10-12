package keeper

import (
	"time"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
	usertokentypes "github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

const (
	maxLiquidationRecords = 1000
)

func (k Keeper) getModuleStats(ctx sdk.Context) types.ModuleStats {
	store := k.getStore(ctx)
	bz := store.Get(types.ModuleStatsKey)
	if bz == nil {
		return types.ModuleStats{}
	}

	var stats types.ModuleStats
	k.cdc.MustUnmarshal(bz, &stats)
	return stats
}

func (k Keeper) setModuleStats(ctx sdk.Context, stats types.ModuleStats) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&stats)
	store.Set(types.ModuleStatsKey, bz)
}

func (k Keeper) getTokenStats(ctx sdk.Context, denom string) types.TokenStats {
	store := k.getStore(ctx)
	bz := store.Get(types.TokenStatsKey(denom))
	if bz == nil {
		return types.TokenStats{Denom: denom}
	}

	var stats types.TokenStats
	k.cdc.MustUnmarshal(bz, &stats)
	if stats.Denom == "" {
		stats.Denom = denom
	}
	return stats
}

func (k Keeper) setTokenStats(ctx sdk.Context, stats types.TokenStats) {
	if stats.Denom == "" {
		return
	}
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&stats)
	store.Set(types.TokenStatsKey(stats.Denom), bz)
}

func (k Keeper) adjustOpenPositions(ctx sdk.Context, delta int64) {
	if delta == 0 {
		return
	}
	stats := k.getModuleStats(ctx)
	if delta > 0 {
		stats.TotalMarginPositions += uint64(delta)
	} else {
		d := uint64(-delta)
		if stats.TotalMarginPositions > d {
			stats.TotalMarginPositions -= d
		} else {
			stats.TotalMarginPositions = 0
		}
	}
	stats.LastUpdated = &time.Time{}
	k.setModuleStats(ctx, stats)
}

func (k Keeper) recordTrade(ctx sdk.Context, denom string, isBuy bool, quoteVolume osmomath.Dec, protocolFee osmomath.Dec, price osmomath.Dec) {
	now := ctx.BlockTime()
	stats := k.getModuleStats(ctx)
	tokenStats := k.getTokenStats(ctx, denom)

	if isBuy {
		stats.TotalBuyVolume = addDecString(stats.TotalBuyVolume, quoteVolume)
		tokenStats.BuyVolume = addDecString(tokenStats.BuyVolume, quoteVolume)
	} else {
		stats.TotalSellVolume = addDecString(stats.TotalSellVolume, quoteVolume)
		tokenStats.SellVolume = addDecString(tokenStats.SellVolume, quoteVolume)
	}

	stats.TotalVolume = addDecString(stats.TotalVolume, quoteVolume)
	stats.ProtocolFeesCollected = addDecString(stats.ProtocolFeesCollected, protocolFee)
	stats.LastUpdated = &now
	k.setModuleStats(ctx, stats)

	tokenStats.TotalVolume = addDecString(tokenStats.TotalVolume, quoteVolume)
	tokenStats.ProtocolFeesCollected = addDecString(tokenStats.ProtocolFeesCollected, protocolFee)
	tokenStats.TradeCount++
	tokenStats.LastPrice = price.String()
	tokenStats.LastTradeAt = &now
	k.setTokenStats(ctx, tokenStats)
}

func (k Keeper) recordLiquidation(ctx sdk.Context, record types.LiquidationRecord) {
	store := k.getStore(ctx)

	seqBz := store.Get(types.LiquidationSeqKey)
	var seq uint64
	if seqBz != nil {
		seq = sdk.BigEndianToUint64(seqBz)
	}
	seq++
	record.Id = seq
	if record.Timestamp.IsZero() {
		record.Timestamp = &time.Time{}
	}

	bz := k.cdc.MustMarshal(&record)
	store.Set(types.LiquidationRecordKey(seq), bz)
	store.Set(types.LiquidationSeqKey, sdk.Uint64ToBigEndian(seq))

	startBz := store.Get(types.LiquidationStartSeqKey)
	var start uint64
	if startBz != nil {
		start = sdk.BigEndianToUint64(startBz)
	}
	if start == 0 {
		start = seq
	}

	if seq-start >= maxLiquidationRecords {
		store.Delete(types.LiquidationRecordKey(start))
		start++
		store.Set(types.LiquidationStartSeqKey, sdk.Uint64ToBigEndian(start))
	} else if startBz == nil {
		store.Set(types.LiquidationStartSeqKey, sdk.Uint64ToBigEndian(start))
	}

	stats := k.getModuleStats(ctx)
	stats.TotalLiquidations++
	stats.LastUpdated = &time.Time{}
	k.setModuleStats(ctx, stats)
}

func (k Keeper) getLiquidationRecords(ctx sdk.Context, denom, trader string, limit, offset uint64) []*types.LiquidationRecord {
	store := k.getStore(ctx)
	start := sdk.BigEndianToUint64(store.Get(types.LiquidationStartSeqKey))
	end := sdk.BigEndianToUint64(store.Get(types.LiquidationSeqKey))
	if end == 0 || limit == 0 {
		return []*types.LiquidationRecord{}
	}

	var records []*types.LiquidationRecord
	skipped := uint64(0)
	for id := end; id >= start && id > 0; id-- {
		var rec types.LiquidationRecord
		bz := store.Get(types.LiquidationRecordKey(id))
		if bz == nil {
			if id == start {
				break
			}
			continue
		}
		k.cdc.MustUnmarshal(bz, &rec)
		if denom != "" && rec.Denom != denom {
			if id == start {
				break
			}
			continue
		}
		if trader != "" && rec.Trader != trader {
			if id == start {
				break
			}
			continue
		}
		if skipped < offset {
			skipped++
			if id == start {
				break
			}
			continue
		}
		copyRec := rec
		records = append(records, &copyRec)
		if uint64(len(records)) >= limit {
			break
		}
		if id == start {
			break
		}
	}
	return records
}

func (k Keeper) setLiquidationRecordsFromGenesis(ctx sdk.Context, records []types.LiquidationRecord, lastSeq uint64) {
	store := k.getStore(ctx)
	if len(records) == 0 {
		if lastSeq > 0 {
			store.Set(types.LiquidationSeqKey, sdk.Uint64ToBigEndian(lastSeq))
		}
		return
	}
	startID := records[0].Id
	for _, record := range records {
		b := k.cdc.MustMarshal(&record)
		store.Set(types.LiquidationRecordKey(record.Id), b)
		if record.Id < startID {
			startID = record.Id
		}
	}
	store.Set(types.LiquidationStartSeqKey, sdk.Uint64ToBigEndian(startID))
	if lastSeq == 0 {
		lastSeq = records[len(records)-1].Id
	}
	store.Set(types.LiquidationSeqKey, sdk.Uint64ToBigEndian(lastSeq))
}

func (k Keeper) getAllTokenStats(ctx sdk.Context) []types.TokenStats {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.TokenStatsKeyPrefix)
	defer iterator.Close()
	stats := make([]types.TokenStats, 0)
	for ; iterator.Valid(); iterator.Next() {
		var ts types.TokenStats
		k.cdc.MustUnmarshal(iterator.Value(), &ts)
		stats = append(stats, ts)
	}
	return stats
}

func (k Keeper) getAllLiquidationRecords(ctx sdk.Context) []types.LiquidationRecord {
	store := k.getStore(ctx)
	start := sdk.BigEndianToUint64(store.Get(types.LiquidationStartSeqKey))
	end := sdk.BigEndianToUint64(store.Get(types.LiquidationSeqKey))
	if end == 0 || start == 0 {
		return []types.LiquidationRecord{}
	}
	records := make([]types.LiquidationRecord, 0, end-start+1)
	for id := start; id <= end; id++ {
		bz := store.Get(types.LiquidationRecordKey(id))
		if bz == nil {
			continue
		}
		var rec types.LiquidationRecord
		k.cdc.MustUnmarshal(bz, &rec)
		records = append(records, rec)
	}
	return records
}

func (k Keeper) countTokens(ctx sdk.Context) uint64 {
	count := uint64(0)
	k.userTokenKeeper.IterateTokens(ctx, func(token usertokentypes.Token) bool {
		count++
		return false
	})
	return count
}

func addDecString(current string, value osmomath.Dec) string {
	if value.IsZero() {
		return current
	}
	total := osmomath.ZeroDec()
	if current != "" {
		existing, err := osmomath.NewDecFromStr(current)
		if err == nil {
			total = existing
		}
	}
	total = total.Add(value)
	return total.String()
}
