package keeper

import (
	"sort"
	"time"

	"cosmossdk.io/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

func (k Keeper) getGlobalPause(ctx sdk.Context) (types.PauseInfo, bool) {
	store := k.getStore(ctx)
	bz := store.Get(types.GlobalPauseKey)
	if bz == nil {
		return types.PauseInfo{}, false
	}
	var info types.PauseInfo
	k.cdc.MustUnmarshal(bz, &info)
	return info, true
}

func (k Keeper) setGlobalPause(ctx sdk.Context, info types.PauseInfo) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&info)
	store.Set(types.GlobalPauseKey, bz)
}

func (k Keeper) getTokenPause(ctx sdk.Context, denom string) (types.PauseInfo, bool) {
	store := prefix.NewStore(k.getStore(ctx), types.TokenPauseKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.PauseInfo{}, false
	}
	var info types.PauseInfo
	k.cdc.MustUnmarshal(bz, &info)
	return info, true
}

func (k Keeper) setTokenPause(ctx sdk.Context, denom string, info types.PauseInfo) {
	store := prefix.NewStore(k.getStore(ctx), types.TokenPauseKeyPrefix)
	bz := k.cdc.MustMarshal(&info)
	store.Set([]byte(denom), bz)
}

func (k Keeper) deleteTokenPause(ctx sdk.Context, denom string) {
	store := prefix.NewStore(k.getStore(ctx), types.TokenPauseKeyPrefix)
	store.Delete([]byte(denom))
}

func (k Keeper) getFreezeInfo(ctx sdk.Context, targetType types.FreezeTargetType, target string) (types.FreezeInfo, bool) {
	store := prefix.NewStore(k.getStore(ctx), types.FreezeKeyPrefix)
	bz := store.Get(types.FreezeKey(targetType, target))
	if bz == nil {
		return types.FreezeInfo{}, false
	}
	var info types.FreezeInfo
	k.cdc.MustUnmarshal(bz, &info)
	return info, true
}

func (k Keeper) setFreezeInfo(ctx sdk.Context, targetType types.FreezeTargetType, target string, info types.FreezeInfo) {
	store := prefix.NewStore(k.getStore(ctx), types.FreezeKeyPrefix)
	bz := k.cdc.MustMarshal(&info)
	store.Set(types.FreezeKey(targetType, target), bz)
}

func (k Keeper) deleteFreezeInfo(ctx sdk.Context, targetType types.FreezeTargetType, target string) {
	store := prefix.NewStore(k.getStore(ctx), types.FreezeKeyPrefix)
	store.Delete(types.FreezeKey(targetType, target))
}

func (k Keeper) getPendingParams(ctx sdk.Context) (types.PendingParams, bool) {
	store := k.getStore(ctx)
	bz := store.Get(types.PendingParamsKey)
	if bz == nil {
		return types.PendingParams{}, false
	}
	var pending types.PendingParams
	k.cdc.MustUnmarshal(bz, &pending)
	return pending, true
}

func (k Keeper) setPendingParams(ctx sdk.Context, pending types.PendingParams) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&pending)
	store.Set(types.PendingParamsKey, bz)
}

func (k Keeper) deletePendingParams(ctx sdk.Context) {
	store := k.getStore(ctx)
	store.Delete(types.PendingParamsKey)
}

func (k Keeper) getEmergencyActionSeq(ctx sdk.Context) uint64 {
	store := k.getStore(ctx)
	bz := store.Get(types.EmergencyActionSeqKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) setEmergencyActionSeq(ctx sdk.Context, seq uint64) {
	store := k.getStore(ctx)
	store.Set(types.EmergencyActionSeqKey, sdk.Uint64ToBigEndian(seq))
}

func (k Keeper) appendEmergencyAction(ctx sdk.Context, action types.EmergencyAction) {
	store := prefix.NewStore(k.getStore(ctx), types.EmergencyActionKeyPrefix)
	bz := k.cdc.MustMarshal(&action)
	store.Set(types.EmergencyActionKey(action.Id), bz)
}

func (k Keeper) setEmergencyAction(ctx sdk.Context, action types.EmergencyAction) {
	store := prefix.NewStore(k.getStore(ctx), types.EmergencyActionKeyPrefix)
	bz := k.cdc.MustMarshal(&action)
	store.Set(types.EmergencyActionKey(action.Id), bz)
}

func (k Keeper) iterateEmergencyActions(ctx sdk.Context, limit uint64, handler func(types.EmergencyAction) bool) {
	store := prefix.NewStore(k.getStore(ctx), types.EmergencyActionKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var count uint64
	for ; iterator.Valid(); iterator.Next() {
		if limit != 0 && count >= limit {
			break
		}

		var action types.EmergencyAction
		k.cdc.MustUnmarshal(iterator.Value(), &action)
		if handler(action) {
			break
		}
		count++
	}
}

func (k Keeper) getEmergencyConfig(ctx sdk.Context) types.EmergencyConfig {
	store := k.getStore(ctx)
	bz := store.Get(types.EmergencyConfigKey)
	if bz == nil {
		return types.EmergencyConfig{
			Signers:   []string{k.authority},
			Threshold: types.DefaultEmergencyThreshold,
		}
	}
	var config types.EmergencyConfig
	k.cdc.MustUnmarshal(bz, &config)
	if len(config.Signers) == 0 {
		config.Signers = []string{k.authority}
	}
	if config.Threshold == 0 {
		config.Threshold = types.DefaultEmergencyThreshold
	}
	return config
}

func (k Keeper) setEmergencyConfig(ctx sdk.Context, config types.EmergencyConfig) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&config)
	store.Set(types.EmergencyConfigKey, bz)
}

func (k Keeper) recordEmergencyAction(ctx sdk.Context, actionType, target, reason string, signers []string) (types.EmergencyAction, error) {
	config := k.getEmergencyConfig(ctx)
	unique := make(map[string]struct{}, len(signers))
	for _, signer := range signers {
		if signer == "" {
			continue
		}
		unique[signer] = struct{}{}
	}

	if uint32(len(unique)) < config.Threshold {
		return types.EmergencyAction{}, types.ErrEmergencyThreshold
	}

	now := ctx.BlockTime().UTC()
	requestedAt := now
	executeAt := now
	nextSeq := k.getEmergencyActionSeq(ctx) + 1
	sorted := sortedStrings(unique)
	action := types.EmergencyAction{
		Id:          nextSeq,
		ActionType:  actionType,
		Target:      target,
		RequestedAt: &requestedAt,
		ExecuteAt:   &executeAt,
		Signers:     sorted,
		Reason:      reason,
		Threshold:   config.Threshold,
	}

	k.setEmergencyAction(ctx, action)
	k.setEmergencyActionSeq(ctx, nextSeq)
	k.updateEmergencyConfig(ctx, sorted)

	return action, nil
}

func sortedStrings(m map[string]struct{}) []string {
	values := make([]string, 0, len(m))
	for key := range m {
		values = append(values, key)
	}
	sort.Strings(values)
	return values
}

func (k Keeper) updateEmergencyConfig(ctx sdk.Context, signers []string) {
	if len(signers) == 0 {
		return
	}

	config := k.getEmergencyConfig(ctx)
	config.Signers = append([]string(nil), signers...)
	if config.Threshold == 0 {
		config.Threshold = types.DefaultEmergencyThreshold
	}
	if uint32(len(config.Signers)) < config.Threshold {
		config.Threshold = uint32(len(config.Signers))
		if config.Threshold == 0 {
			config.Threshold = types.DefaultEmergencyThreshold
		}
	}
	now := ctx.BlockTime().UTC()
	config.UpdatedAt = &now
	k.setEmergencyConfig(ctx, config)
}

func (k Keeper) verifyAuthority(msgAuthority string) error {
	if msgAuthority != k.authority {
		return types.ErrUnauthorizedAuthority
	}
	return nil
}

func (k Keeper) buildPauseInfo(paused bool, reason string, resumeIn *time.Duration, updatedAt time.Time) types.PauseInfo {
	info := types.PauseInfo{
		Paused: paused,
		Reason: reason,
	}
	var resumeAt *time.Time
	if resumeIn != nil && *resumeIn > 0 {
		t := updatedAt.Add(*resumeIn).UTC()
		resumeAt = &t
	}
	return info.WithTimestamps(resumeAt, updatedAt)
}

func (k Keeper) buildFreezeInfo(frozen bool, reason string, duration *time.Duration, updatedAt time.Time) types.FreezeInfo {
	info := types.FreezeInfo{
		Frozen: frozen,
		Reason: reason,
	}
	var unfreezeAt *time.Time
	if duration != nil && *duration > 0 {
		t := updatedAt.Add(*duration).UTC()
		unfreezeAt = &t
	}
	return info.WithTimestamps(unfreezeAt, updatedAt)
}
