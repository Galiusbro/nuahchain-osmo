package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Params returns the module parameters
func (k queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Position returns a specific position by ID
func (k queryServer) Position(goCtx context.Context, req *types.QueryPositionRequest) (*types.QueryPositionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.PositionId == "" {
		return nil, status.Error(codes.InvalidArgument, "position ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	position, found := k.GetPosition(ctx, req.PositionId)
	if !found {
		return nil, status.Error(codes.NotFound, "position not found")
	}

	// Update PnL with current price
	if position.Status == types.PositionStatusOpen {
		currentPrice, err := k.GetTokenPrice(ctx, position.TokenDenom)
		if err == nil {
			k.UpdatePositionPnL(ctx, &position, currentPrice)
		}
	}

	return &types.QueryPositionResponse{Position: position}, nil
}

// Positions returns all positions, optionally filtered
func (k queryServer) Positions(goCtx context.Context, req *types.QueryPositionsRequest) (*types.QueryPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	positionStore := prefix.NewStore(store, types.PositionKeyPrefix)

	var positions []types.Position
	pageRes, err := query.Paginate(positionStore, req.Pagination, func(key []byte, value []byte) error {
		var position types.Position
		if err := k.cdc.Unmarshal(value, &position); err != nil {
			return err
		}

		// Apply filters
		if req.Status != types.PositionStatusUnspecified && position.Status != req.Status {
			return nil
		}

		if req.TokenDenom != "" && position.TokenDenom != req.TokenDenom {
			return nil
		}

		// Update PnL with current price for open positions
		if position.Status == types.PositionStatusOpen {
			currentPrice, err := k.GetTokenPrice(ctx, position.TokenDenom)
			if err == nil {
				k.UpdatePositionPnL(ctx, &position, currentPrice)
			}
		}

		positions = append(positions, position)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPositionsResponse{
		Positions:  positions,
		Pagination: pageRes,
	}, nil
}

// PositionsByTrader returns all positions for a specific trader
func (k queryServer) PositionsByTrader(goCtx context.Context, req *types.QueryPositionsByTraderRequest) (*types.QueryPositionsByTraderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Trader == "" {
		return nil, status.Error(codes.InvalidArgument, "trader address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	traderStore := prefix.NewStore(store, types.TraderPositionPrefix(req.Trader))

	var positions []types.Position
	pageRes, err := query.Paginate(traderStore, req.Pagination, func(key []byte, value []byte) error {
		positionID := string(value)
		position, found := k.GetPosition(ctx, positionID)
		if !found {
			return nil // Skip if position not found
		}

		// Apply status filter
		if req.Status != types.PositionStatusUnspecified && position.Status != req.Status {
			return nil
		}

		// Update PnL with current price for open positions
		if position.Status == types.PositionStatusOpen {
			currentPrice, err := k.GetTokenPrice(ctx, position.TokenDenom)
			if err == nil {
				k.UpdatePositionPnL(ctx, &position, currentPrice)
			}
		}

		positions = append(positions, position)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPositionsByTraderResponse{
		Positions:  positions,
		Pagination: pageRes,
	}, nil
}

// LiquidationPrice calculates the liquidation price for given parameters
func (k queryServer) LiquidationPrice(goCtx context.Context, req *types.QueryLiquidationPriceRequest) (*types.QueryLiquidationPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.CollateralAmount.IsNil() || req.CollateralAmount.LTE(math.ZeroInt()) {
		return nil, status.Error(codes.InvalidArgument, "collateral amount must be positive")
	}

	if req.PositionSize.IsNil() || req.PositionSize.LTE(math.ZeroInt()) {
		return nil, status.Error(codes.InvalidArgument, "position size must be positive")
	}

	if req.EntryPrice.IsNil() || req.EntryPrice.LTE(math.LegacyZeroDec()) {
		return nil, status.Error(codes.InvalidArgument, "entry price must be positive")
	}

	if req.Side == types.PositionSideUnspecified {
		return nil, status.Error(codes.InvalidArgument, "position side must be specified")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	liquidationPrice := k.CalculateLiquidationPrice(ctx, req.CollateralAmount, req.PositionSize, req.EntryPrice, req.Side)

	return &types.QueryLiquidationPriceResponse{
		LiquidationPrice: liquidationPrice,
	}, nil
}

// EstimatePosition estimates the outcome of opening a position with given parameters
func (k queryServer) EstimatePosition(goCtx context.Context, req *types.QueryEstimatePositionRequest) (*types.QueryEstimatePositionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.TokenDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "token denom cannot be empty")
	}

	if req.CollateralAmount.IsNil() || req.CollateralAmount.LTE(math.ZeroInt()) {
		return nil, status.Error(codes.InvalidArgument, "collateral amount must be positive")
	}

	if req.Leverage.IsNil() || req.Leverage.LTE(math.LegacyOneDec()) {
		return nil, status.Error(codes.InvalidArgument, "leverage must be greater than 1")
	}

	if req.Side == types.PositionSideUnspecified {
		return nil, status.Error(codes.InvalidArgument, "position side must be specified")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate token
	if !k.ValidateTokenDenom(ctx, req.TokenDenom) {
		return nil, status.Error(codes.InvalidArgument, "token not supported for leverage trading")
	}

	// Get current token price
	currentPrice, err := k.GetTokenPrice(ctx, req.TokenDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get token price: %v", err))
	}

	// Calculate position size
	collateralDec := math.LegacyNewDecFromInt(req.CollateralAmount)
	positionValue := collateralDec.Mul(req.Leverage)
	positionSize := positionValue.Quo(currentPrice).TruncateInt()

	// Calculate liquidation price
	liquidationPrice := k.CalculateLiquidationPrice(ctx, req.CollateralAmount, positionSize, currentPrice, req.Side)

	// Calculate trading fee
	params := k.GetParams(ctx)
	tradingFee := positionValue.Mul(params.TradingFee).TruncateInt()

	return &types.QueryEstimatePositionResponse{
		PositionSize:     positionSize,
		EntryPrice:       currentPrice,
		LiquidationPrice: liquidationPrice,
		TradingFee:       tradingFee,
	}, nil
}

// TokenPrice returns the current price of a token from the usertoken bonding curve
func (k queryServer) TokenPrice(goCtx context.Context, req *types.QueryTokenPriceRequest) (*types.QueryTokenPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "denom cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate token
	if !k.ValidateTokenDenom(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "token not supported for leverage trading")
	}

	// Get current price
	price, err := k.GetTokenPrice(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get token price: %v", err))
	}

	// Get current supply
	supply, err := k.userTokenKeeper.GetTokenSupply(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get token supply: %v", err))
	}

	return &types.QueryTokenPriceResponse{
		Price:  price,
		Supply: supply,
	}, nil
}

// LendingPools returns all lending pools
func (k queryServer) LendingPools(goCtx context.Context, req *types.QueryLendingPoolsRequest) (*types.QueryLendingPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pools := k.lendingKeeper.GetAllLendingPools(ctx)

	return &types.QueryLendingPoolsResponse{Pools: pools}, nil
}

// LendingPool returns a specific lending pool by denom
func (k queryServer) LendingPool(goCtx context.Context, req *types.QueryLendingPoolRequest) (*types.QueryLendingPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "denom cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pool, exists := k.lendingKeeper.GetLendingPool(ctx, req.Denom)
	if !exists {
		return nil, status.Error(codes.NotFound, "lending pool not found")
	}

	return &types.QueryLendingPoolResponse{Pool: pool}, nil
}

// BorrowPositions returns all borrow positions for a borrower
func (k queryServer) BorrowPositions(goCtx context.Context, req *types.QueryBorrowPositionsRequest) (*types.QueryBorrowPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Borrower == "" {
		return nil, status.Error(codes.InvalidArgument, "borrower cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	positions := k.lendingKeeper.GetAllBorrowPositions(ctx)

	// Filter positions by borrower
	var borrowerPositions []types.BorrowPosition
	for _, pos := range positions {
		if pos.Borrower == req.Borrower {
			borrowerPositions = append(borrowerPositions, pos)
		}
	}

	return &types.QueryBorrowPositionsResponse{Positions: borrowerPositions}, nil
}

// LiquidityProviders returns all liquidity providers
func (k queryServer) LiquidityProviders(goCtx context.Context, req *types.QueryLiquidityProvidersRequest) (*types.QueryLiquidityProvidersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	providers := k.lendingKeeper.GetAllLiquidityProviders(ctx)

	return &types.QueryLiquidityProvidersResponse{Providers: providers}, nil
}

// Stats returns leverage module statistics
func (k queryServer) Stats(goCtx context.Context, req *types.QueryStatsRequest) (*types.QueryStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all positions
	allPositions := k.GetAllPositions(ctx)

	// Get all lending pools
	allPools := k.lendingKeeper.GetAllLendingPools(ctx)

	// Get all borrow positions
	allBorrowPositions := k.lendingKeeper.GetAllBorrowPositions(ctx)

	// Get all liquidity providers
	allProviders := k.lendingKeeper.GetAllLiquidityProviders(ctx)

	return &types.QueryStatsResponse{
		TotalPositions:          uint64(len(allPositions)),
		TotalLendingPools:       uint64(len(allPools)),
		TotalBorrowPositions:    uint64(len(allBorrowPositions)),
		TotalLiquidityProviders: uint64(len(allProviders)),
	}, nil
}
