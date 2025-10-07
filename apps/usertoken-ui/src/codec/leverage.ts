// Leverage trading message types and interfaces

export const MSG_OPEN_POSITION_TYPE_URL = "/osmosis.leverage.v1beta1.MsgOpenPosition";
export const MSG_CLOSE_POSITION_TYPE_URL = "/osmosis.leverage.v1beta1.MsgClosePosition";
export const MSG_ADD_COLLATERAL_TYPE_URL = "/osmosis.leverage.v1beta1.MsgAddCollateral";
export const MSG_REMOVE_COLLATERAL_TYPE_URL = "/osmosis.leverage.v1beta1.MsgRemoveCollateral";
export const MSG_LIQUIDATE_POSITION_TYPE_URL = "/osmosis.leverage.v1beta1.MsgLiquidatePosition";
export const MSG_PROVIDE_LIQUIDITY_TYPE_URL = "/osmosis.leverage.v1beta1.MsgProvideLiquidity";

// Position side enum
export enum PositionSide {
    UNSPECIFIED = 0,
    LONG = 1,
    SHORT = 2
}

// Position status enum
export enum PositionStatus {
    UNSPECIFIED = 0,
    OPEN = 1,
    CLOSED = 2,
    LIQUIDATED = 3
}

// Position interface
export interface Position {
    id: string;
    trader: string;
    tokenDenom: string;
    collateralDenom: string;
    side: PositionSide;
    size: string;
    collateral: string;
    leverage: string;
    entryPrice: string;
    liquidationPrice: string;
    unrealizedPnl: string;
    status: PositionStatus;
    createdAt: number;
    updatedAt: number;
}

// Leverage parameters interface
export interface LeverageParams {
    maxLeverage: string;
    maintenanceMargin: string;
    liquidationFee: string;
    tradingFee: string;
    maxPositionSize: string;
    minCollateralAmount: string;
    baseInterestRate: string;
    interestRateMultiplier: string;
    maxInterestRate: string;
    maxBorrowRatio: string;
}

// Lending pool interface
export interface LendingPool {
    denom: string;
    totalSupply: string;
    totalBorrowed: string;
    availableLiquidity: string;
    interestRate: string;
    utilizationRate: string;
    lastUpdateTime: number;
}

// Borrow position interface
export interface BorrowPosition {
    id: string;
    borrower: string;
    tokenDenom: string;
    borrowedAmount: string;
    accruedInterest: string;
    interestRate: string;
    createdAt: number;
    lastInterestTime: number;
    leveragePositionId: string;
}

// Liquidity provider interface
export interface LiquidityProvider {
    provider: string;
    tokenDenom: string;
    amount: string;
    shareTokens: string;
    providedAt: number;
}

// Message interfaces for protobuf
export interface MsgOpenPosition {
    trader: string;
    tokenDenom: string;
    collateral: {
        denom: string;
        amount: string;
    };
    leverage: string;
    side: PositionSide;
    minPrice: string;
    maxPrice: string;
}

export interface MsgClosePosition {
    trader: string;
    positionId: string;
    minPrice: string;
    maxPrice: string;
}

export interface MsgAddCollateral {
    trader: string;
    positionId: string;
    amount: {
        denom: string;
        amount: string;
    };
}

export interface MsgRemoveCollateral {
    trader: string;
    positionId: string;
    amount: {
        denom: string;
        amount: string;
    };
}

export interface MsgLiquidatePosition {
    liquidator: string;
    positionId: string;
}

export interface MsgProvideLiquidity {
    provider: string;
    amount: {
        denom: string;
        amount: string;
    };
}

// Response interfaces
export interface MsgOpenPositionResponse {
    positionId: string;
    position: Position;
}

export interface MsgClosePositionResponse {
    realizedPnl: string;
    collateralReturned: {
        denom: string;
        amount: string;
    };
}

export interface MsgAddCollateralResponse {
    newLiquidationPrice: string;
}

export interface MsgRemoveCollateralResponse {
    newLiquidationPrice: string;
}

export interface MsgLiquidatePositionResponse {
    liquidationReward: {
        denom: string;
        amount: string;
    };
    remainingCollateral: {
        denom: string;
        amount: string;
    };
}

export interface MsgProvideLiquidityResponse {
    shareTokens: string;
}

// Query interfaces
export interface QueryPositionRequest {
    positionId: string;
}

export interface QueryPositionResponse {
    position: Position;
}

export interface QueryPositionsByTraderRequest {
    trader: string;
    pagination?: {
        key?: string;
        offset?: number;
        limit?: number;
        countTotal?: boolean;
        reverse?: boolean;
    };
    status?: PositionStatus;
}

export interface QueryPositionsByTraderResponse {
    positions: Position[];
    pagination?: {
        nextKey?: string;
        total?: number;
    };
}

export interface QueryEstimatePositionRequest {
    tokenDenom: string;
    collateralAmount: string;
    leverage: string;
    side: PositionSide;
}

export interface QueryEstimatePositionResponse {
    positionSize: string;
    entryPrice: string;
    liquidationPrice: string;
    tradingFee: string;
}

export interface QueryTokenPriceRequest {
    denom: string;
}

export interface QueryTokenPriceResponse {
    price: string;
    supply: string;
}

export interface QueryLiquidationPriceRequest {
    collateralAmount: string;
    positionSize: string;
    entryPrice: string;
    side: PositionSide;
}

export interface QueryLiquidationPriceResponse {
    liquidationPrice: string;
}

// Additional query interfaces
export interface QueryLendingPoolsResponse {
    pools: LendingPool[];
    pagination?: {
        nextKey?: string;
        total?: number;
    };
}

export interface QueryLendingPoolResponse {
    pool: LendingPool;
}

export interface QueryBorrowPositionsByBorrowerResponse {
    positions: BorrowPosition[];
    pagination?: {
        nextKey?: string;
        total?: number;
    };
}

export interface QueryLiquidityProvidersResponse {
    providers: LiquidityProvider[];
    pagination?: {
        nextKey?: string;
        total?: number;
    };
}

export interface QueryParamsResponse {
    params: LeverageParams;
}

export interface QueryStatsResponse {
    totalPositions: number;
    totalLendingPools: number;
    totalBorrowPositions: number;
    totalLiquidityProviders: number;
}

// Helper functions for position side
export function positionSideToString(side: PositionSide): string {
    switch (side) {
        case PositionSide.LONG:
            return "LONG";
        case PositionSide.SHORT:
            return "SHORT";
        default:
            return "UNSPECIFIED";
    }
}

export function stringToPositionSide(side: string): PositionSide {
    switch (side.toUpperCase()) {
        case "LONG":
            return PositionSide.LONG;
        case "SHORT":
            return PositionSide.SHORT;
        default:
            return PositionSide.UNSPECIFIED;
    }
}

// Helper functions for position status
export function positionStatusToString(status: PositionStatus): string {
    switch (status) {
        case PositionStatus.OPEN:
            return "OPEN";
        case PositionStatus.CLOSED:
            return "CLOSED";
        case PositionStatus.LIQUIDATED:
            return "LIQUIDATED";
        default:
            return "UNSPECIFIED";
    }
}

export function stringToPositionStatus(status: string): PositionStatus {
    switch (status.toUpperCase()) {
        case "OPEN":
            return PositionStatus.OPEN;
        case "CLOSED":
            return PositionStatus.CLOSED;
        case "LIQUIDATED":
            return PositionStatus.LIQUIDATED;
        default:
            return PositionStatus.UNSPECIFIED;
    }
}
