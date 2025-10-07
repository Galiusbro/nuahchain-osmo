import { Registry } from "@cosmjs/proto-signing";
import { assertIsDeliverTxSuccess, DeliverTxResponse } from "@cosmjs/stargate";
import Decimal from "decimal.js";
import { useCallback, useMemo, useRef, useState } from 'react';
import {
    BorrowPosition,
    LendingPool,
    LeverageParams,
    Position,
    PositionSide,
    QueryBorrowPositionsByBorrowerResponse,
    QueryLendingPoolResponse,
    QueryLendingPoolsResponse,
    QueryParamsResponse,
    QueryPositionResponse,
    QueryPositionsByTraderResponse,
    QueryStatsResponse,
    QueryTokenPriceResponse
} from '../codec/leverage';
import {
    MsgAddCollateral,
    MsgClosePosition,
    MsgLiquidatePosition,
    MsgOpenPosition,
    MsgProvideLiquidity,
    MsgRemoveCollateral
} from '../codec/leverage_proto';

interface UseLeverageProps {
    walletAddress: string;
    restEndpoint: string;
    registry: Registry;
    signAndBroadcast: (messages: any[], fee: any, memo?: string) => Promise<DeliverTxResponse>;
}

const LEGACY_DECIMAL_PLACES = 18;
const LEGACY_MULTIPLIER = new Decimal(10).pow(LEGACY_DECIMAL_PLACES);

const toLegacyDecString = (input: string): string => {
    const value = input?.trim() ?? "";
    if (!value) {
        return "0";
    }

    const dec = new Decimal(value);
    if (!dec.isFinite()) {
        throw new Error(`Invalid decimal value: ${input}`);
    }

    return dec.mul(LEGACY_MULTIPLIER).toFixed(0, Decimal.ROUND_DOWN);
};

const BASE64_REGEX = /^[A-Za-z0-9+/]+={0,2}$/;

const maybeDecodeBase64 = (value?: string): string => {
    if (!value) {
        return '';
    }

    const normalized = value.trim();
    if (!normalized || normalized.length % 4 !== 0 || !BASE64_REGEX.test(normalized)) {
        return value;
    }

    try {
        if (typeof atob === 'function') {
            return atob(normalized);
        }
    } catch {
        // ignore and try other strategies
    }

    return value;
};

const extractPositionIdFromRawLog = (rawLog?: string): string | undefined => {
    if (!rawLog) {
        return undefined;
    }

    try {
        const parsed = JSON.parse(rawLog);
        if (Array.isArray(parsed)) {
            for (const log of parsed) {
                const events = Array.isArray(log?.events) ? log.events : [];
                for (const event of events) {
                    const eventType = maybeDecodeBase64(event?.type);
                    if (eventType !== 'open_position' && eventType !== 'leverage.position_created') {
                        continue;
                    }
                    const attributes = Array.isArray(event?.attributes) ? event.attributes : [];
                    for (const attribute of attributes) {
                        const key = maybeDecodeBase64(attribute?.key);
                        if (key === 'position_id') {
                            return maybeDecodeBase64(attribute?.value);
                        }
                    }
                }
            }
        }
    } catch {
        // ignore, fall back to regex search below
    }

    const match = rawLog.match(/position_id["=:\s]*"?([0-9a-zA-Z_-]+)"?/);
    return match?.[1];
};

export function useLeverage({ walletAddress, restEndpoint, registry, signAndBroadcast }: UseLeverageProps) {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string>('');

    // Cache for preventing duplicate requests
    const cacheRef = useRef<{
        params?: LeverageParams;
        positions?: Position[];
        pools?: LendingPool[];
        lastFetch?: number;
    }>({});

    // Helper function for retry logic
    const fetchWithRetry = useCallback(async (url: string, retries = 2): Promise<Response> => {
        for (let i = 0; i < retries; i++) {
            try {
                const response = await fetch(url);
                if (response.ok) {
                    return response;
                }
                if (i === retries - 1) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
            } catch (err) {
                if (i === retries - 1) {
                    throw err;
                }
                // Wait before retry (shorter delay)
                await new Promise(resolve => setTimeout(resolve, 500 * (i + 1)));
            }
        }
        throw new Error('Max retries exceeded');
    }, []);

    // Query functions
    const queryParams = useCallback(async (): Promise<LeverageParams> => {
        // Check cache first
        if (cacheRef.current.params && cacheRef.current.lastFetch && Date.now() - cacheRef.current.lastFetch < 30000) {
            return cacheRef.current.params;
        }

        try {
            const response = await fetchWithRetry(`${restEndpoint}/osmosis/leverage/v1beta1/params`);
            const data: QueryParamsResponse = await response.json();
            cacheRef.current.params = data.params;
            cacheRef.current.lastFetch = Date.now();
            return data.params;
        } catch (err) {
            console.error('Failed to query leverage params:', err);
            throw err;
        }
    }, [restEndpoint, fetchWithRetry]);

    const queryPositions = useCallback(async (): Promise<Position[]> => {
        if (!walletAddress) return [];

        try {
            const response = await fetchWithRetry(`${restEndpoint}/osmosis/leverage/v1beta1/positions/trader/${walletAddress}`);
            const data: QueryPositionsByTraderResponse = await response.json();
            return data.positions || [];
        } catch (err) {
            console.error('Failed to query positions:', err);
            throw err;
        }
    }, [walletAddress, restEndpoint, fetchWithRetry]);

    const queryPosition = useCallback(async (positionId: string): Promise<Position> => {
        try {
            const response = await fetch(`${restEndpoint}/osmosis/leverage/v1beta1/position/${positionId}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data: QueryPositionResponse = await response.json();
            return data.position;
        } catch (err) {
            console.error('Failed to query position:', err);
            throw err;
        }
    }, [restEndpoint]);

    const queryTokenPrice = useCallback(async (denom: string): Promise<string> => {
        try {
            const response = await fetch(`${restEndpoint}/osmosis/leverage/v1beta1/token-price/${encodeURIComponent(denom)}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data: QueryTokenPriceResponse = await response.json();
            return data.price;
        } catch (err) {
            console.error('Failed to query token price:', err);
            throw err;
        }
    }, [restEndpoint]);

    const queryLendingPools = useCallback(async (): Promise<LendingPool[]> => {
        // Check cache first
        if (cacheRef.current.pools && cacheRef.current.lastFetch && Date.now() - cacheRef.current.lastFetch < 30000) {
            return cacheRef.current.pools;
        }

        try {
            const response = await fetchWithRetry(`${restEndpoint}/osmosis/leverage/v1beta1/lending-pools`);
            const data: QueryLendingPoolsResponse = await response.json();
            const pools = data.pools || [];
            cacheRef.current.pools = pools;
            cacheRef.current.lastFetch = Date.now();
            return pools;
        } catch (err) {
            console.error('Failed to query lending pools:', err);
            throw err;
        }
    }, [restEndpoint, fetchWithRetry]);

    const queryLendingPool = useCallback(async (denom: string): Promise<LendingPool | null> => {
        try {
            const response = await fetch(`${restEndpoint}/osmosis/leverage/v1beta1/lending-pool/${encodeURIComponent(denom)}`);
            if (!response.ok) {
                return null;
            }
            const data: QueryLendingPoolResponse = await response.json();
            return data.pool;
        } catch (err) {
            console.error('Failed to query lending pool:', err);
            return null;
        }
    }, [restEndpoint]);

    const queryBorrowPositions = useCallback(async (): Promise<BorrowPosition[]> => {
        if (!walletAddress) return [];

        try {
            const response = await fetch(`${restEndpoint}/osmosis/leverage/v1beta1/borrow-positions/borrower/${walletAddress}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data: QueryBorrowPositionsByBorrowerResponse = await response.json();
            return data.positions || [];
        } catch (err) {
            console.error('Failed to query borrow positions:', err);
            throw err;
        }
    }, [walletAddress, restEndpoint]);

    const queryStats = useCallback(async (): Promise<QueryStatsResponse> => {
        try {
            const response = await fetch(`${restEndpoint}/osmosis/leverage/v1beta1/stats`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            return await response.json();
        } catch (err) {
            console.error('Failed to query stats:', err);
            throw err;
        }
    }, [restEndpoint]);

    // Transaction functions
    const openPosition = useCallback(async (params: {
        tokenDenom: string;
        collateralAmount: string;
        collateralDenom: string;
        leverage: string;
        side: PositionSide;
        minPrice: string;
        maxPrice: string;
    }): Promise<string> => {
        setLoading(true);
        setError('');

        try {
            const msg = {
                typeUrl: "/osmosis.leverage.v1beta1.MsgOpenPosition",
                value: MsgOpenPosition.fromPartial({
                    trader: walletAddress,
                    token_denom: params.tokenDenom,
                    collateral: {
                        denom: params.collateralDenom,
                        amount: params.collateralAmount
                    },
                    leverage: toLegacyDecString(params.leverage),
                    side: params.side,
                    min_price: toLegacyDecString(params.minPrice),
                    max_price: toLegacyDecString(params.maxPrice)
                })
            };

            const fee = {
                amount: [{ denom: "unuah", amount: "1000" }],
                gas: "500000"
            };

            const result = await signAndBroadcast([msg], fee, "Open leverage position");
            assertIsDeliverTxSuccess(result);

            const events = result.events ?? [];

            const findPositionIdInEvents = (): string | undefined => {
                for (const event of events) {
                    const eventType = maybeDecodeBase64(event.type);
                    if (eventType !== 'open_position' && eventType !== 'leverage.position_created') {
                        continue;
                    }

                    for (const attribute of event.attributes ?? []) {
                        const key = maybeDecodeBase64(attribute.key);
                        if (key === 'position_id') {
                            return maybeDecodeBase64(attribute.value);
                        }
                    }
                }

                for (const event of events) {
                    for (const attribute of event.attributes ?? []) {
                        const key = maybeDecodeBase64(attribute.key);
                        if (key === 'position_id') {
                            return maybeDecodeBase64(attribute.value);
                        }
                    }
                }

                return undefined;
            };

            const positionId = findPositionIdInEvents() ?? extractPositionIdFromRawLog(result.rawLog);

            if (!positionId) {
                throw new Error("Position ID not found in transaction events");
            }

            return positionId;
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to open position';
            setError(errorMsg);
            throw err;
        } finally {
            setLoading(false);
        }
    }, [walletAddress, signAndBroadcast]);

    const closePosition = useCallback(async (positionId: string, minPrice: string, maxPrice: string): Promise<void> => {
        setLoading(true);
        setError('');

        try {
            const msg = {
                typeUrl: "/osmosis.leverage.v1beta1.MsgClosePosition",
                value: MsgClosePosition.fromPartial({
                    trader: walletAddress,
                    position_id: positionId,
                    min_price: toLegacyDecString(minPrice),
                    max_price: toLegacyDecString(maxPrice)
                })
            };

            const fee = {
                amount: [{ denom: "unuah", amount: "1000" }],
                gas: "500000"
            };

            const result = await signAndBroadcast([msg], fee, "Close leverage position");
            assertIsDeliverTxSuccess(result);
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to close position';
            setError(errorMsg);
            throw err;
        } finally {
            setLoading(false);
        }
    }, [walletAddress, signAndBroadcast]);

    const addCollateral = useCallback(async (positionId: string, amount: string, denom: string): Promise<void> => {
        setLoading(true);
        setError('');

        try {
            const msg = {
                typeUrl: "/osmosis.leverage.v1beta1.MsgAddCollateral",
                value: MsgAddCollateral.fromPartial({
                    trader: walletAddress,
                    position_id: positionId,
                    amount: {
                        denom,
                        amount
                    }
                })
            };

            const fee = {
                amount: [{ denom: "unuah", amount: "1000" }],
                gas: "200000"
            };

            const result = await signAndBroadcast([msg], fee, "Add collateral to position");
            assertIsDeliverTxSuccess(result);
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to add collateral';
            setError(errorMsg);
            throw err;
        } finally {
            setLoading(false);
        }
    }, [walletAddress, signAndBroadcast]);

    const removeCollateral = useCallback(async (positionId: string, amount: string, denom: string): Promise<void> => {
        setLoading(true);
        setError('');

        try {
            const msg = {
                typeUrl: "/osmosis.leverage.v1beta1.MsgRemoveCollateral",
                value: MsgRemoveCollateral.fromPartial({
                    trader: walletAddress,
                    position_id: positionId,
                    amount: {
                        denom,
                        amount
                    }
                })
            };

            const fee = {
                amount: [{ denom: "unuah", amount: "1000" }],
                gas: "200000"
            };

            const result = await signAndBroadcast([msg], fee, "Remove collateral from position");
            assertIsDeliverTxSuccess(result);
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to remove collateral';
            setError(errorMsg);
            throw err;
        } finally {
            setLoading(false);
        }
    }, [walletAddress, signAndBroadcast]);

    const liquidatePosition = useCallback(async (positionId: string): Promise<void> => {
        setLoading(true);
        setError('');

        try {
            const msg = {
                typeUrl: "/osmosis.leverage.v1beta1.MsgLiquidatePosition",
                value: MsgLiquidatePosition.fromPartial({
                    liquidator: walletAddress,
                    position_id: positionId
                })
            };

            const fee = {
                amount: [{ denom: "unuah", amount: "1000" }],
                gas: "500000"
            };

            const result = await signAndBroadcast([msg], fee, "Liquidate position");
            assertIsDeliverTxSuccess(result);
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to liquidate position';
            setError(errorMsg);
            throw err;
        } finally {
            setLoading(false);
        }
    }, [walletAddress, signAndBroadcast]);

    const provideLiquidity = useCallback(async (amount: string, denom: string): Promise<void> => {
        setLoading(true);
        setError('');

        try {
            const msg = {
                typeUrl: "/osmosis.leverage.v1beta1.MsgProvideLiquidity",
                value: MsgProvideLiquidity.fromPartial({
                    provider: walletAddress,
                    amount: {
                        denom,
                        amount
                    }
                })
            };

            const fee = {
                amount: [{ denom: "unuah", amount: "1000" }],
                gas: "500000"
            };

            const result = await signAndBroadcast([msg], fee, "Provide liquidity to lending pool");
            assertIsDeliverTxSuccess(result);
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Failed to provide liquidity';
            setError(errorMsg);
            throw err;
        } finally {
            setLoading(false);
        }
    }, [walletAddress, signAndBroadcast]);

    // Utility functions
    const formatAmount = useCallback((amount: string | undefined, decimals: number = 6): string => {
        if (!amount || amount === 'undefined' || amount === 'null') {
            return '0';
        }
        try {
            return new Decimal(amount).div(Math.pow(10, decimals)).toFixed(decimals);
        } catch (error) {
            console.warn('Error formatting amount:', amount, error);
            return '0';
        }
    }, []);

    const formatPrice = useCallback((price: string | undefined): string => {
        if (!price || price === 'undefined' || price === 'null') {
            return '0';
        }
        try {
            return new Decimal(price).div(1_000_000).toFixed(6);
        } catch (error) {
            console.warn('Error formatting price:', price, error);
            return '0';
        }
    }, []);

    const calculatePositionValue = useCallback((size: string | undefined, price: string | undefined): string => {
        if (!size || !price || size === 'undefined' || price === 'undefined') {
            return '0';
        }
        try {
            return new Decimal(size).mul(price).toString();
        } catch (error) {
            console.warn('Error calculating position value:', size, price, error);
            return '0';
        }
    }, []);

    const calculateLiquidationPrice = useCallback((
        entryPrice: string | undefined,
        leverage: string | undefined,
        maintenanceMargin: string | undefined,
        side: PositionSide
    ): string => {
        if (!entryPrice || !leverage || !maintenanceMargin ||
            entryPrice === 'undefined' || leverage === 'undefined' || maintenanceMargin === 'undefined') {
            return '0';
        }

        try {
            const entry = new Decimal(entryPrice);
            const lev = new Decimal(leverage);
            const margin = new Decimal(maintenanceMargin);

            if (side === PositionSide.LONG) {
                return entry.mul(new Decimal(1).sub(margin)).div(lev).toString();
            } else {
                return entry.mul(new Decimal(1).add(margin)).div(lev).toString();
            }
        } catch (error) {
            console.warn('Error calculating liquidation price:', entryPrice, leverage, maintenanceMargin, error);
            return '0';
        }
    }, []);

    return useMemo(() => ({
        // State
        loading,
        error,
        setError,

        // Query functions
        queryParams,
        queryPositions,
        queryPosition,
        queryTokenPrice,
        queryLendingPools,
        queryLendingPool,
        queryBorrowPositions,
        queryStats,

        // Transaction functions
        openPosition,
        closePosition,
        addCollateral,
        removeCollateral,
        liquidatePosition,
        provideLiquidity,

        // Utility functions
        formatAmount,
        formatPrice,
        calculatePositionValue,
        calculateLiquidationPrice
    }), [
        loading,
        error,
        setError,
        queryParams,
        queryPositions,
        queryPosition,
        queryTokenPrice,
        queryLendingPools,
        queryLendingPool,
        queryBorrowPositions,
        queryStats,
        openPosition,
        closePosition,
        addCollateral,
        removeCollateral,
        liquidatePosition,
        provideLiquidity,
        formatAmount,
        formatPrice,
        calculatePositionValue,
        calculateLiquidationPrice
    ]);
}
