package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammtypes "github.com/osmosis-labs/osmosis/v30/x/gamm/types"
	pooltypes "github.com/osmosis-labs/osmosis/v30/x/poolmanager/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	IterateAllDenomMetaData(ctx sdk.Context, cb func(banktypes.Metadata) bool)
}

// PoolKeeper defines the expected interface for pool management.
type PoolKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (pooltypes.PoolI, error)
	GetPools(ctx sdk.Context) ([]pooltypes.PoolI, error)
	GetPoolsWithWildcard(ctx sdk.Context, poolType string) ([]pooltypes.PoolI, error)
	CreatePool(ctx sdk.Context, msg pooltypes.CreatePoolMsg) (uint64, error)
	GetNextPoolId(ctx sdk.Context) uint64
	SetNextPoolId(ctx sdk.Context, poolId uint64)
	GetTotalLiquidity(ctx sdk.Context) sdk.Coins
	GetPoolDenoms(ctx sdk.Context, poolId uint64) ([]string, error)
	CalculateSpotPrice(ctx sdk.Context, poolId uint64, quoteAssetDenom, baseAssetDenom string) (osmomath.BigDec, error)
	RouteExactAmountIn(ctx sdk.Context, sender sdk.AccAddress, routes []pooltypes.SwapAmountInRoute, tokenIn sdk.Coin, tokenOutMinAmount osmomath.Int) (tokenOutAmount osmomath.Int, err error)
	RouteExactAmountOut(ctx sdk.Context, sender sdk.AccAddress, routes []pooltypes.SwapAmountOutRoute, tokenInMaxAmount osmomath.Int, tokenOut sdk.Coin) (tokenInAmount osmomath.Int, err error)
	MultihopEstimateOutGivenExactAmountIn(ctx sdk.Context, routes []pooltypes.SwapAmountInRoute, tokenIn sdk.Coin) (tokenOutAmount osmomath.Int, err error)
	MultihopEstimateInGivenExactAmountOut(ctx sdk.Context, routes []pooltypes.SwapAmountOutRoute, tokenOut sdk.Coin) (tokenInAmount osmomath.Int, err error)
	GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error)
}

// GAMMKeeper defines the expected interface for GAMM module interactions.
type GAMMKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (gammtypes.CFMMPoolI, error)
	GetPools(ctx sdk.Context) []gammtypes.CFMMPoolI
	GetPoolsAndPoke(ctx sdk.Context) []gammtypes.CFMMPoolI
	GetPoolAndPoke(ctx sdk.Context, poolId uint64) (gammtypes.CFMMPoolI, error)
	CreatePool(ctx sdk.Context, poolI gammtypes.CFMMPoolI) (uint64, error)
	GetNextPoolId(ctx sdk.Context) uint64
	SetNextPoolId(ctx sdk.Context, poolId uint64)
	CleanupBalancerPool(ctx sdk.Context, poolId uint64, denom0, denom1 string) error
	GetPoolType(ctx sdk.Context, poolId uint64) (string, error)
	GetPoolsWithWildcard(ctx sdk.Context, poolType string) ([]gammtypes.CFMMPoolI, error)
	GetCFMMPool(ctx sdk.Context, poolId uint64) (gammtypes.CFMMPoolI, error)
	GetPoolDenoms(ctx sdk.Context, poolId uint64) ([]string, error)
	CalculateSpotPrice(ctx sdk.Context, poolId uint64, quoteAssetDenom, baseAssetDenom string) (osmomath.BigDec, error)
	JoinPoolNoSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareOutAmount osmomath.Int, tokenInMaxs sdk.Coins) (tokenIn sdk.Coins, sharesOut osmomath.Int, err error)
	JoinSwapExactAmountIn(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, tokenIn sdk.Coin, shareOutMinAmount osmomath.Int) (sharesOut osmomath.Int, err error)
	JoinSwapShareAmountOut(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, tokenInDenom string, shareOutAmount osmomath.Int, tokenInMaxAmount osmomath.Int) (tokenInAmount osmomath.Int, err error)
	ExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount osmomath.Int, tokenOutMins sdk.Coins) (tokenOut sdk.Coins, err error)
	ExitSwapShareAmountIn(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, tokenOutDenom string, shareInAmount osmomath.Int, tokenOutMinAmount osmomath.Int) (tokenOutAmount osmomath.Int, err error)
	ExitSwapExactAmountOut(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, tokenOut sdk.Coin, shareInMaxAmount osmomath.Int) (shareInAmount osmomath.Int, err error)
	SwapExactAmountIn(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, tokenIn sdk.Coin, tokenOutDenom string, tokenOutMinAmount osmomath.Int) (tokenOutAmount osmomath.Int, err error)
	SwapExactAmountOut(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, tokenInDenom string, tokenInMaxAmount osmomath.Int, tokenOut sdk.Coin) (tokenInAmount osmomath.Int, err error)
	GetTotalLiquidity(ctx sdk.Context) sdk.Coins
	GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error)
}

// EpochsKeeper defines the expected interface for epochs module interactions.
type EpochsKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
	SetEpochInfo(ctx sdk.Context, epoch epochstypes.EpochInfo)
	DeleteEpochInfo(ctx sdk.Context, identifier string)
	IterateEpochInfo(ctx sdk.Context, fn func(index int64, epochInfo epochstypes.EpochInfo) (stop bool))
	GetAllEpochInfos(ctx sdk.Context) []epochstypes.EpochInfo
}

// GovKeeper defines the expected interface for governance module interactions.
type GovKeeper interface {
	GetProposal(ctx sdk.Context, proposalID uint64) (govtypes.Proposal, bool)
	SetProposal(ctx sdk.Context, proposal govtypes.Proposal)
	GetProposals(ctx sdk.Context) (proposals govtypes.Proposals)
	GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (govtypes.Vote, bool)
	SetVote(ctx sdk.Context, vote govtypes.Vote)
	GetVotes(ctx sdk.Context, proposalID uint64) (votes govtypes.Votes)
	GetDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) (govtypes.Deposit, bool)
	SetDeposit(ctx sdk.Context, deposit govtypes.Deposit)
	GetDeposits(ctx sdk.Context, proposalID uint64) (deposits govtypes.Deposits)
	GetParams(ctx sdk.Context) govtypes.Params
	SetParams(ctx sdk.Context, params govtypes.Params)
	GetDepositParams(ctx sdk.Context) govtypes.DepositParams
	GetVotingParams(ctx sdk.Context) govtypes.VotingParams
	GetTallyParams(ctx sdk.Context) govtypes.TallyParams
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) sdk.ModuleAccountI
}

// DistrKeeper defines the expected distribution keeper used for simulations (noalias)
type DistrKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
	DistributeFromFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error
}

// StakingKeeper defines the expected staking keeper used for simulations (noalias)
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	GetAllValidators(ctx sdk.Context) []interface{} // Using interface{} to avoid import cycles
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (interface{}, bool)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (interface{}, bool)
	GetAllDelegations(ctx sdk.Context) []interface{}
	GetUnbondingDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (interface{}, bool)
	GetAllUnbondingDelegations(ctx sdk.Context, delAddr sdk.AccAddress) []interface{}
	GetRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) (interface{}, bool)
	GetAllRedelegations(ctx sdk.Context, delAddr sdk.AccAddress, srcValAddress, dstValAddress sdk.ValAddress) []interface{}
}

// TWAPKeeper defines the expected interface for TWAP (Time-Weighted Average Price) calculations.
type TWAPKeeper interface {
	GetArithmeticTwap(ctx sdk.Context, poolId uint64, baseAssetDenom, quoteAssetDenom string, startTime, endTime *time.Time) (osmomath.Dec, error)
	GetArithmeticTwapToNow(ctx sdk.Context, poolId uint64, baseAssetDenom, quoteAssetDenom string, startTime time.Time) (osmomath.Dec, error)
	GetGeometricTwap(ctx sdk.Context, poolId uint64, baseAssetDenom, quoteAssetDenom string, startTime, endTime *time.Time) (osmomath.Dec, error)
	GetGeometricTwapToNow(ctx sdk.Context, poolId uint64, baseAssetDenom, quoteAssetDenom string, startTime time.Time) (osmomath.Dec, error)
	GetBeginBlockAccumulatorValue(ctx sdk.Context, poolId uint64, baseAssetDenom, quoteAssetDenom string) (osmomath.Dec, error)
	GetMostRecentRecordStoreRepresentation(ctx sdk.Context, poolId uint64, baseAssetDenom, quoteAssetDenom string) (interface{}, error)
	GetRecordAtOrBeforeTime(ctx sdk.Context, poolId uint64, timestamp time.Time, baseAssetDenom, quoteAssetDenom string) (interface{}, error)
	GetAllMostRecentRecordsForPool(ctx sdk.Context, poolId uint64) ([]interface{}, error)
	GetAllHistoricalPoolIndexedTWAPs(ctx sdk.Context) ([]interface{}, error)
	GetAllHistoricalTimeIndexedTWAPs(ctx sdk.Context) ([]interface{}, error)
}

// ConcentratedLiquidityKeeper defines the expected interface for concentrated liquidity interactions.
type ConcentratedLiquidityKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (interface{}, error)
	GetPools(ctx sdk.Context) ([]interface{}, error)
	CreateConcentratedPool(ctx sdk.Context, msg interface{}) (uint64, error)
	GetPosition(ctx sdk.Context, positionId uint64) (interface{}, error)
	GetUserPositions(ctx sdk.Context, addr sdk.AccAddress, poolId uint64) ([]interface{}, error)
	CreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, tokensProvided sdk.Coins, tokenMinAmount0, tokenMinAmount1 osmomath.Int, lowerTick, upperTick int64) (interface{}, error)
	WithdrawPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, requestedLiquidityAmountToWithdraw osmomath.Dec) (amtDenom0, amtDenom1 osmomath.Int, err error)
	AddToPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, amount0Added, amount1Added osmomath.Int) (uint64, osmomath.Int, osmomath.Int, error)
	CollectSpreadRewards(ctx sdk.Context, owner sdk.AccAddress, positionId uint64) (sdk.Coins, error)
	CollectIncentives(ctx sdk.Context, owner sdk.AccAddress, positionId uint64) (sdk.Coins, sdk.Coins, error)
	GetTotalLiquidity(ctx sdk.Context) sdk.Coins
	GetConcentratedPoolById(ctx sdk.Context, poolId uint64) (interface{}, error)
}

// IncentivesKeeper defines the expected interface for incentives module interactions.
type IncentivesKeeper interface {
	GetGauges(ctx sdk.Context) []interface{}
	GetActiveGauges(ctx sdk.Context) []interface{}
	GetUpcomingGauges(ctx sdk.Context) []interface{}
	GetGaugeByID(ctx sdk.Context, gaugeID uint64) (interface{}, error)
	CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo interface{}, startTime time.Time, numEpochsPaidOver uint64) (uint64, error)
	AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error
	Distribute(ctx sdk.Context, gauges []interface{}) (sdk.Coins, error)
	GetParams(ctx sdk.Context) interface{}
	SetParams(ctx sdk.Context, params interface{})
}
