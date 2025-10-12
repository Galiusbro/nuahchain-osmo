package keeper

import (
	"context"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
	usertokentypes "github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
)

type queryServer struct {
	Keeper
}

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the QueryServer interface.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{Keeper: k}
}

func (s queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := s.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (s queryServer) GlobalPause(goCtx context.Context, req *types.QueryGlobalPauseRequest) (*types.QueryGlobalPauseResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	info, found := s.getGlobalPause(ctx)
	if !found {
		return &types.QueryGlobalPauseResponse{}, nil
	}
	return &types.QueryGlobalPauseResponse{Status: &info}, nil
}

func (s queryServer) TokenPause(goCtx context.Context, req *types.QueryTokenPauseRequest) (*types.QueryTokenPauseResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	info, found := s.getTokenPause(ctx, req.Denom)
	if !found {
		return &types.QueryTokenPauseResponse{}, nil
	}
	return &types.QueryTokenPauseResponse{Status: &info}, nil
}

func (s queryServer) Freeze(goCtx context.Context, req *types.QueryFreezeRequest) (*types.QueryFreezeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.TargetType == types.FreezeTargetType_FREEZE_TARGET_TYPE_UNSPECIFIED || req.Target == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid freeze target")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	info, found := s.getFreezeInfo(ctx, req.TargetType, req.Target)
	if !found {
		return &types.QueryFreezeResponse{}, nil
	}
	return &types.QueryFreezeResponse{Status: &info}, nil
}

func (s queryServer) PendingParams(goCtx context.Context, req *types.QueryPendingParamsRequest) (*types.QueryPendingParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	pending, found := s.getPendingParams(ctx)
	if !found {
		return &types.QueryPendingParamsResponse{}, nil
	}
	return &types.QueryPendingParamsResponse{PendingParams: &pending}, nil
}

func (s queryServer) EmergencyActions(goCtx context.Context, req *types.QueryEmergencyActionsRequest) (*types.QueryEmergencyActionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	actions := make([]*types.EmergencyAction, 0)
	limit := req.Limit
	s.iterateEmergencyActions(ctx, limit, func(action types.EmergencyAction) bool {
		actionCopy := action
		actions = append(actions, &actionCopy)
		return false
	})
	return &types.QueryEmergencyActionsResponse{Actions: actions}, nil
}

func (s queryServer) EmergencyConfig(goCtx context.Context, req *types.QueryEmergencyConfigRequest) (*types.QueryEmergencyConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	config := s.getEmergencyConfig(ctx)
	return &types.QueryEmergencyConfigResponse{Config: &config}, nil
}

func (s queryServer) ListTokens(goCtx context.Context, req *types.QueryListTokensRequest) (*types.QueryListTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}
	var tokens []*types.TokenSummary
	collected := uint64(0)
	skipped := uint64(0)

	s.userTokenKeeper.IterateTokens(ctx, func(token usertokentypes.Token) bool {
		if req.Offset > 0 && skipped < req.Offset {
			skipped++
			return false
		}
		if collected >= limit {
			return true
		}
		stats := s.getTokenStats(ctx, token.Denom)
		summary := types.TokenSummary{
			Denom:          token.Denom,
			Name:           token.Name,
			Symbol:         token.Symbol,
			Creator:        token.Creator,
			CurveCompleted: token.State.CurveCompleted,
			CurrentPrice:   token.State.CurrentPrice,
			TokensSold:     token.State.BondingCurveSold,
			Stats:          &stats,
		}
		if summary.CurrentPrice == "" {
			if pool, found := s.GetPool(ctx, token.Denom); found {
				summary.CurrentPrice = pool.LastPrice
			}
		}
		if summary.Stats.LastPrice == "" {
			summary.Stats.LastPrice = summary.CurrentPrice
		}
		// ensure stats denom populated
		if summary.Stats.Denom == "" {
			summary.Stats.Denom = token.Denom
		}
		copySummary := summary
		tokens = append(tokens, &copySummary)
		collected++
		return false
	})

	return &types.QueryListTokensResponse{Tokens: tokens}, nil
}

func (s queryServer) TokenStats(goCtx context.Context, req *types.QueryTokenStatsRequest) (*types.QueryTokenStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if strings.TrimSpace(req.Denom) == "" {
		return nil, status.Error(codes.InvalidArgument, "denom is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	stats := s.getTokenStats(ctx, req.Denom)
	if stats.LastPrice == "" {
		if pool, found := s.GetPool(ctx, req.Denom); found {
			stats.LastPrice = pool.LastPrice
		}
	}
	if stats.Denom == "" {
		stats.Denom = req.Denom
	}
	return &types.QueryTokenStatsResponse{Stats: &stats}, nil
}

func (s queryServer) ModuleStats(goCtx context.Context, req *types.QueryModuleStatsRequest) (*types.QueryModuleStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	stats := s.getModuleStats(ctx)
	stats.TotalTokensCreated = s.countTokens(ctx)
	if stats.LastUpdated.IsZero() {
		stats.LastUpdated = &time.Time{}
	}
	return &types.QueryModuleStatsResponse{Stats: &stats}, nil
}

func (s queryServer) MarginPositions(goCtx context.Context, req *types.QueryMarginPositionsRequest) (*types.QueryMarginPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}
	offset := req.Offset

	positions := make([]*types.MarginPosition, 0)
	summary := types.MarginSummary{}
	summary.Denom = req.Denom

	skipped := uint64(0)
	collected := uint64(0)

	store := s.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.MarginPositionKeyPrefix)
	defer iterator.Close()

	totalCollateral := osmomath.ZeroDec()
	totalLong := osmomath.ZeroDec()
	totalShort := osmomath.ZeroDec()
	openPositions := uint64(0)

	for ; iterator.Valid(); iterator.Next() {
		var position types.MarginPosition
		s.cdc.MustUnmarshal(iterator.Value(), &position)
		if req.Denom != "" && position.Denom != req.Denom {
			continue
		}
		if req.Trader != "" && position.Trader != req.Trader {
			continue
		}
		if position.Status != types.PositionStatus_POSITION_STATUS_OPEN {
			continue
		}
		collateral := types.MustNewDecFromStr(position.CollateralAmount)
		exposure := types.MustNewDecFromStr(position.PositionSize)
		totalCollateral = totalCollateral.Add(collateral)
		if position.Type == types.PositionType_POSITION_TYPE_LONG {
			totalLong = totalLong.Add(exposure)
		} else {
			totalShort = totalShort.Add(exposure)
		}
		openPositions++
		if offset > 0 && skipped < offset {
			skipped++
			continue
		}
		if collected >= limit {
			continue
		}
		copyPos := position
		positions = append(positions, &copyPos)
		collected++
	}

	summary.TotalCollateral = totalCollateral.String()
	summary.TotalLongExposure = totalLong.String()
	summary.TotalShortExposure = totalShort.String()
	summary.OpenPositions = openPositions

	return &types.QueryMarginPositionsResponse{
		Positions: positions,
		Summary:   &summary,
	}, nil
}

func (s queryServer) Liquidations(goCtx context.Context, req *types.QueryLiquidationsRequest) (*types.QueryLiquidationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	limit := req.Limit
	if limit == 0 {
		limit = 50
	}
	records := s.getLiquidationRecords(ctx, req.Denom, req.Trader, limit, req.Offset)
	return &types.QueryLiquidationsResponse{Records: records}, nil
}
