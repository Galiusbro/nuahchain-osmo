import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import Decimal from "decimal.js";
import { AlertTriangle, Coins, Minus, Plus, RefreshCw, TrendingDown, TrendingUp } from 'lucide-react';
import type { FormEvent } from "react";
import { useCallback, useEffect, useState } from "react";
import {
    LendingPool,
    LeverageParams,
    Position,
    PositionSide,
    positionSideToString,
    positionStatusToString
} from '../codec/leverage';
import { useLeverage } from '../hooks/useLeverage';
import { useTokens } from '../hooks/useTokens';

interface LeverageFormProps {
    walletAddress: string;
    restEndpoint: string;
    registry: any;
    signAndBroadcast: (messages: any[], fee: any, memo?: string) => Promise<any>;
}

export default function LeverageForm({
    walletAddress,
    restEndpoint,
    registry,
    signAndBroadcast
}: LeverageFormProps) {
    const { allTokens: tokens, isFetchingTokens: tokensLoading, fetchTokens } = useTokens({ restBaseUrl: restEndpoint, walletAddress });
    const leverage = useLeverage({ walletAddress, restEndpoint, registry, signAndBroadcast });

    // Load data on mount
    useEffect(() => {
        if (!restEndpoint || !walletAddress) return;

        setLeverageLoading(true);

        const loadData = async () => {
            try {
                // Load all data in parallel for better performance
                const [paramsResult, positionsResult, poolsResult] = await Promise.allSettled([
                    leverage.queryParams(),
                    leverage.queryPositions(),
                    leverage.queryLendingPools()
                ]);

                // Set results if successful
                if (paramsResult.status === 'fulfilled') {
                    setParams(paramsResult.value);
                }
                if (positionsResult.status === 'fulfilled') {
                    setPositions(positionsResult.value);
                }
                if (poolsResult.status === 'fulfilled') {
                    setLendingPools(poolsResult.value);
                }

                // Load tokens separately to avoid conflicts
                await fetchTokens();
            } catch (error) {
                console.error('Failed to load data:', error);
            } finally {
                setLeverageLoading(false);
            }
        };

        loadData();
    }, [restEndpoint, walletAddress]);

    // State
    const [positions, setPositions] = useState<Position[]>([]);
    const [lendingPools, setLendingPools] = useState<LendingPool[]>([]);
    const [params, setParams] = useState<LeverageParams | null>(null);
    const [leverageLoading, setLeverageLoading] = useState(false);
    const [selectedToken, setSelectedToken] = useState<string>('');
    const [collateralAmount, setCollateralAmount] = useState<string>('');
    const [leverageAmount, setLeverageAmount] = useState<string>('2');
    const [side, setSide] = useState<PositionSide>(PositionSide.LONG);
    const [tokenPrice, setTokenPrice] = useState<string>('0');
    const [liquidityAmount, setLiquidityAmount] = useState<string>('');
    const [liquidityDenom, setLiquidityDenom] = useState<string>('unuah');
    const [collateralToAdd, setCollateralToAdd] = useState<string>('');
    const [collateralToRemove, setCollateralToRemove] = useState<string>('');

    // Load token price
    const loadTokenPrice = useCallback(async (denom: string) => {
        if (!denom) return;
        try {
            const price = await leverage.queryTokenPrice(denom);
            // Convert from micro-units to normal units (divide by 1,000,000)
            const priceInNormalUnits = new Decimal(price).div(1_000_000).toString();
            setTokenPrice(priceInNormalUnits);
        } catch (err) {
            console.error('Failed to load token price:', err);
        }
    }, [leverage]);

    useEffect(() => {
        loadTokenPrice(selectedToken);
    }, [selectedToken, loadTokenPrice]);

    // Form handlers
    const handleOpenPosition = async (e: FormEvent) => {
        e.preventDefault();
        if (!selectedToken || !collateralAmount || !leverageAmount) return;

        try {
            const collateralAmountMicro = new Decimal(collateralAmount).mul(1_000_000).toFixed(0);

            // Convert UI price (NUAH) into micro units for slippage checks
            const priceMicro = tokenPrice && tokenPrice !== '0'
                ? new Decimal(tokenPrice).mul(1_000_000)
                : null;

            const minPrice = side === PositionSide.SHORT && priceMicro
                ? priceMicro.mul(0.9).toFixed(18)
                : '0';

            const maxPrice = side === PositionSide.LONG && priceMicro
                ? priceMicro.mul(1.1).toFixed(18)
                : '999999999000000';

            const leverageValue = leverageAmount; // Send as simple string like "2"
            console.log('Opening position with params:', {
                tokenDenom: selectedToken,
                collateralAmount: collateralAmountMicro,
                collateralDenom: 'unuah',
                leverage: leverageValue,
                side,
                minPrice,
                maxPrice
            });

            await leverage.openPosition({
                tokenDenom: selectedToken,
                collateralAmount: collateralAmountMicro,
                collateralDenom: 'unuah',
                leverage: leverageValue,
                side,
                minPrice,
                maxPrice
            });

            // Reload data
            const [positionsData, poolsData, paramsData] = await Promise.allSettled([
                leverage.queryPositions(),
                leverage.queryLendingPools(),
                leverage.queryParams()
            ]);
            if (positionsData.status === 'fulfilled') setPositions(positionsData.value);
            if (poolsData.status === 'fulfilled') setLendingPools(poolsData.value);
            if (paramsData.status === 'fulfilled') setParams(paramsData.value);

            // Reset form
            setCollateralAmount('');
            setLeverageAmount('2');
        } catch (err) {
            console.error('Failed to open position:', err);
        }
    };

    const handleClosePosition = async (positionId: string) => {
        try {
            const minPrice = '0';
            const maxPrice = '999999999000000'; // Micro units
            await leverage.closePosition(positionId, minPrice, maxPrice);
            // Reload positions
            const positionsData = await leverage.queryPositions();
            setPositions(positionsData);
        } catch (err) {
            console.error('Failed to close position:', err);
        }
    };

    const handleAddCollateral = async (positionId: string) => {
        if (!collateralToAdd) return;
        try {
            const amountMicro = new Decimal(collateralToAdd).mul(1_000_000).toString();
            await leverage.addCollateral(positionId, amountMicro, 'unuah');
            // Reload positions
            const positionsData = await leverage.queryPositions();
            setPositions(positionsData);
            setCollateralToAdd('');
        } catch (err) {
            console.error('Failed to add collateral:', err);
        }
    };

    const handleRemoveCollateral = async (positionId: string) => {
        if (!collateralToRemove) return;
        try {
            const amountMicro = new Decimal(collateralToRemove).mul(1_000_000).toString();
            await leverage.removeCollateral(positionId, amountMicro, 'unuah');
            // Reload positions
            const positionsData = await leverage.queryPositions();
            setPositions(positionsData);
            setCollateralToRemove('');
        } catch (err) {
            console.error('Failed to remove collateral:', err);
        }
    };

    const handleProvideLiquidity = async (e: FormEvent) => {
        e.preventDefault();
        if (!liquidityAmount) return;

        try {
            const amountMicro = new Decimal(liquidityAmount).mul(1_000_000).toString();
            await leverage.provideLiquidity(amountMicro, liquidityDenom);
            // Reload lending pools
            const poolsData = await leverage.queryLendingPools();
            setLendingPools(poolsData);
            setLiquidityAmount('');
        } catch (err) {
            console.error('Failed to provide liquidity:', err);
        }
    };

    // Calculate position estimates
    const calculatePositionSize = () => {
        if (!collateralAmount || !leverageAmount || !tokenPrice || tokenPrice === '0') return '0';

        try {
            const collateral = new Decimal(collateralAmount);
            const leverage = new Decimal(leverageAmount);
            const price = new Decimal(tokenPrice);
            const positionValue = collateral.mul(leverage);
            return positionValue.div(price).toFixed(6);
        } catch (error) {
            console.warn('Error calculating position size:', error);
            return '0';
        }
    };

    const calculateLiquidationPrice = () => {
        if (!tokenPrice || !leverageAmount || !params || !params.maintenanceMargin) return '0';

        try {
            const price = new Decimal(tokenPrice);
            const leverage = new Decimal(leverageAmount);
            const margin = new Decimal(params.maintenanceMargin);

            if (side === PositionSide.LONG) {
                return price.mul(new Decimal(1).sub(margin)).div(leverage).toFixed(6);
            } else {
                return price.mul(new Decimal(1).add(margin)).div(leverage).toFixed(6);
            }
        } catch (error) {
            console.warn('Error calculating liquidation price:', error);
            return '0';
        }
    };

    if (!walletAddress) {
        return (
            <Card className="mb-8">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <TrendingUp className="h-5 w-5" />
                        🚀 Leverage Trading (100x)
                    </CardTitle>
                    <CardDescription>
                        Connect your wallet to access leverage trading
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <p className="text-muted-foreground text-center py-8">
                        Please connect your wallet first to access leverage trading.
                    </p>
                </CardContent>
            </Card>
        );
    }

    if (leverageLoading) {
        return (
            <Card className="mb-8">
                <CardContent className="flex items-center justify-center py-8">
                    <div className="text-center">
                        <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-4" />
                        <p className="text-muted-foreground">Loading leverage data...</p>
                    </div>
                </CardContent>
            </Card>
        );
    }

    return (
        <div className="space-y-6">
            {/* Risk Warning */}
            <Card className="border-yellow-200 bg-yellow-50 dark:bg-yellow-950/20">
                <CardContent className="pt-6">
                    <div className="flex items-start gap-3">
                        <AlertTriangle className="h-5 w-5 text-yellow-600 mt-0.5" />
                        <div>
                            <h3 className="font-semibold text-yellow-800 dark:text-yellow-200">
                                High Risk Warning
                            </h3>
                            <p className="text-sm text-yellow-700 dark:text-yellow-300 mt-1">
                                Leverage trading involves significant risk. You can lose more than your initial investment.
                                Only trade with money you can afford to lose.
                            </p>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <Tabs defaultValue="open" className="space-y-6">
                <TabsList className="grid w-full grid-cols-4">
                    <TabsTrigger value="open">Open Position</TabsTrigger>
                    <TabsTrigger value="positions">My Positions</TabsTrigger>
                    <TabsTrigger value="liquidity">Liquidity</TabsTrigger>
                    <TabsTrigger value="pools">Lending Pools</TabsTrigger>
                </TabsList>

                {/* Open Position Tab */}
                <TabsContent value="open">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                {side === PositionSide.LONG ? (
                                    <TrendingUp className="h-5 w-5 text-green-600" />
                                ) : (
                                    <TrendingDown className="h-5 w-5 text-red-600" />
                                )}
                                Open {positionSideToString(side)} Position
                            </CardTitle>
                            <CardDescription>
                                Create a new leverage position with up to 100x leverage
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <form onSubmit={handleOpenPosition} className="space-y-6">
                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="token">Token</Label>
                                        <select
                                            id="token"
                                            value={selectedToken}
                                            onChange={(e) => setSelectedToken(e.target.value)}
                                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                            required
                                            disabled={tokensLoading}
                                        >
                                            <option value="">
                                                {tokensLoading ? "Loading tokens..." : "Select token..."}
                                            </option>
                                            {tokens?.map((token) => (
                                                <option key={token.denom} value={token.denom}>
                                                    {token.symbol || token.denom.split('/').pop()}
                                                </option>
                                            )) || []}
                                        </select>
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="collateral">Collateral (NUAH)</Label>
                                        <Input
                                            id="collateral"
                                            type="number"
                                            value={collateralAmount}
                                            onChange={(e) => setCollateralAmount(e.target.value)}
                                            placeholder="100"
                                            min="0"
                                            step="0.000001"
                                            required
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="leverage">Leverage (1x - 100x)</Label>
                                        <Input
                                            id="leverage"
                                            type="number"
                                            value={leverageAmount}
                                            onChange={(e) => setLeverageAmount(e.target.value)}
                                            min="1"
                                            max="100"
                                            step="0.1"
                                            required
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="side">Position Side</Label>
                                        <select
                                            id="side"
                                            value={side}
                                            onChange={(e) => setSide(parseInt(e.target.value) as PositionSide)}
                                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                        >
                                            <option value={PositionSide.LONG}>LONG (Buy)</option>
                                            <option value={PositionSide.SHORT}>SHORT (Sell)</option>
                                        </select>
                                    </div>
                                </div>

                                {/* Position Estimate */}
                                {selectedToken && collateralAmount && leverageAmount && tokenPrice !== '0' && (
                                    <Card className="bg-blue-50 dark:bg-blue-950/20">
                                        <CardHeader>
                                            <CardTitle className="text-lg">📊 Position Estimate</CardTitle>
                                        </CardHeader>
                                        <CardContent className="space-y-2">
                                            <div className="flex justify-between">
                                                <span className="font-medium">Position Size:</span>
                                                <span>{calculatePositionSize()} tokens</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span className="font-medium">Entry Price:</span>
                                                <span>{leverage.formatPrice(tokenPrice)} NUAH</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span className="font-medium">Liquidation Price:</span>
                                                <span>{calculateLiquidationPrice()} NUAH</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span className="font-medium">Required Margin:</span>
                                                <span>{collateralAmount} NUAH</span>
                                            </div>
                                        </CardContent>
                                    </Card>
                                )}

                                <Button
                                    type="submit"
                                    disabled={leverage.loading || !selectedToken || !collateralAmount}
                                    className={`w-full ${side === PositionSide.LONG
                                        ? 'bg-green-600 hover:bg-green-700'
                                        : 'bg-red-600 hover:bg-red-700'
                                        }`}
                                >
                                    {leverage.loading ? (
                                        <>
                                            <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                                            Opening...
                                        </>
                                    ) : (
                                        <>
                                            {side === PositionSide.LONG ? (
                                                <TrendingUp className="h-4 w-4 mr-2" />
                                            ) : (
                                                <TrendingDown className="h-4 w-4 mr-2" />
                                            )}
                                            Open {positionSideToString(side)} Position
                                        </>
                                    )}
                                </Button>
                            </form>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* My Positions Tab */}
                <TabsContent value="positions">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                📋 My Positions
                            </CardTitle>
                            <CardDescription>
                                Manage your active leverage positions
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {positions.length === 0 ? (
                                <div className="text-center py-8">
                                    <p className="text-muted-foreground">
                                        No positions found. Open your first leverage position above!
                                    </p>
                                </div>
                            ) : (
                                <div className="space-y-4">
                                    {positions.map((position) => {
                                        const pnlRaw = position.unrealizedPnl ?? '0';
                                        const pnlValue = pnlRaw === 'undefined' || pnlRaw === 'null' ? '0' : pnlRaw;
                                        const pnlDecimal = new Decimal(pnlValue || '0');

                                        return (
                                            <Card key={position.id} className="border-l-4 border-l-blue-500">
                                            <CardContent className="pt-6">
                                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                                                    <div>
                                                        <Label className="text-sm font-medium text-muted-foreground">Position</Label>
                                                        <p className="font-mono text-sm">#{position.id}</p>
                                                        <p className={`text-sm font-bold ${position.side === PositionSide.LONG ? 'text-green-600' : 'text-red-600'}`}>
                                                            {positionSideToString(position.side)}
                                                        </p>
                                                    </div>
                                                    <div>
                                                        <Label className="text-sm font-medium text-muted-foreground">Size</Label>
                                                        <p className="text-sm">{leverage.formatAmount(position.size)} tokens</p>
                                                        <p className="text-sm">{leverage.formatAmount(position.collateral)} NUAH collateral</p>
                                                    </div>
                                                    <div>
                                                        <Label className="text-sm font-medium text-muted-foreground">Prices</Label>
                                                        <p className="text-sm">Entry: {leverage.formatPrice(position.entryPrice)}</p>
                                                        <p className="text-sm">Liquidation: {leverage.formatPrice(position.liquidationPrice)}</p>
                                                    </div>
                                                    <div>
                                                        <Label className="text-sm font-medium text-muted-foreground">PnL & Status</Label>
                                                        <p className={`text-sm font-bold ${pnlDecimal.isPositive() ? 'text-green-600' : 'text-red-600'}`}>
                                                            {pnlDecimal.isPositive() ? '+' : ''}
                                                            {leverage.formatAmount(pnlValue)} NUAH
                                                        </p>
                                                        <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${position.status === 1 ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                                                            }`}>
                                                            {positionStatusToString(position.status)}
                                                        </span>
                                                    </div>
                                                </div>

                                                {position.status === 1 && (
                                                    <div className="mt-4 pt-4 border-t space-y-4">
                                                        <div className="flex gap-2">
                                                            <Input
                                                                placeholder="Add collateral (NUAH)"
                                                                value={collateralToAdd}
                                                                onChange={(e) => setCollateralToAdd(e.target.value)}
                                                                className="flex-1"
                                                            />
                                                            <Button
                                                                onClick={() => handleAddCollateral(position.id)}
                                                                disabled={!collateralToAdd || leverage.loading}
                                                                size="sm"
                                                            >
                                                                <Plus className="h-4 w-4 mr-1" />
                                                                Add
                                                            </Button>
                                                        </div>
                                                        <div className="flex gap-2">
                                                            <Input
                                                                placeholder="Remove collateral (NUAH)"
                                                                value={collateralToRemove}
                                                                onChange={(e) => setCollateralToRemove(e.target.value)}
                                                                className="flex-1"
                                                            />
                                                            <Button
                                                                onClick={() => handleRemoveCollateral(position.id)}
                                                                disabled={!collateralToRemove || leverage.loading}
                                                                size="sm"
                                                                variant="outline"
                                                            >
                                                                <Minus className="h-4 w-4 mr-1" />
                                                                Remove
                                                            </Button>
                                                        </div>
                                                        <Button
                                                            onClick={() => handleClosePosition(position.id)}
                                                            disabled={leverage.loading}
                                                            className="w-full"
                                                            variant="destructive"
                                                        >
                                                            Close Position
                                                        </Button>
                                                    </div>
                                                )}
                                            </CardContent>
                                            </Card>
                                        );
                                    })}
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* Provide Liquidity Tab */}
                <TabsContent value="liquidity">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Coins className="h-5 w-5" />
                                Provide Liquidity
                            </CardTitle>
                            <CardDescription>
                                Add liquidity to lending pools to earn interest
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <form onSubmit={handleProvideLiquidity} className="space-y-4">
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="liquidityAmount">Amount</Label>
                                        <Input
                                            id="liquidityAmount"
                                            type="number"
                                            value={liquidityAmount}
                                            onChange={(e) => setLiquidityAmount(e.target.value)}
                                            placeholder="1000"
                                            min="0"
                                            step="0.000001"
                                            required
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <Label htmlFor="liquidityDenom">Denomination</Label>
                                        <select
                                            id="liquidityDenom"
                                            value={liquidityDenom}
                                            onChange={(e) => setLiquidityDenom(e.target.value)}
                                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                        >
                                            <option value="unuah">UNUAH</option>
                                        </select>
                                    </div>
                                </div>
                                <Button
                                    type="submit"
                                    disabled={leverage.loading || !liquidityAmount}
                                    className="w-full"
                                >
                                    {leverage.loading ? (
                                        <>
                                            <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                                            Providing...
                                        </>
                                    ) : (
                                        <>
                                            <Coins className="h-4 w-4 mr-2" />
                                            Provide Liquidity
                                        </>
                                    )}
                                </Button>
                            </form>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* Lending Pools Tab */}
                <TabsContent value="pools">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                🏦 Lending Pools
                            </CardTitle>
                            <CardDescription>
                                View available lending pools and their statistics
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {lendingPools.length === 0 ? (
                                <div className="text-center py-8">
                                    <p className="text-muted-foreground">
                                        No lending pools found. Provide liquidity to create pools!
                                    </p>
                                </div>
                            ) : (
                                <div className="space-y-4">
                                    {lendingPools.map((pool) => (
                                        <Card key={pool.denom} className="border-l-4 border-l-purple-500">
                                            <CardContent className="pt-6">
                                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                                                    <div>
                                                        <Label className="text-sm font-medium text-muted-foreground">Token</Label>
                                                        <p className="font-mono text-sm">{pool.denom.split('/').pop()}</p>
                                                    </div>
                                                    <div>
                                                        <Label className="text-sm font-medium text-muted-foreground">Total Supply</Label>
                                                        <p className="text-sm">{leverage.formatAmount(pool.totalSupply)}</p>
                                                    </div>
                                                    <div>
                                                        <Label className="text-sm font-medium text-muted-foreground">Total Borrowed</Label>
                                                        <p className="text-sm">{leverage.formatAmount(pool.totalBorrowed)}</p>
                                                    </div>
                                                    <div>
                                                        <Label className="text-sm font-medium text-muted-foreground">Available Liquidity</Label>
                                                        <p className="text-sm">{leverage.formatAmount(pool.availableLiquidity)}</p>
                                                    </div>
                                                    <div>
                                                        <Label className="text-sm font-medium text-muted-foreground">Interest Rate</Label>
                                                        <p className="text-sm font-bold text-green-600">
                                                            {pool.interestRate ? new Decimal(pool.interestRate).mul(100).toFixed(2) : '0.00'}%
                                                        </p>
                                                    </div>
                                                </div>
                                            </CardContent>
                                        </Card>
                                    ))}
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>

            {leverage.error && (
                <Card className="border-red-200 bg-red-50 dark:bg-red-950/20">
                    <CardContent className="pt-6">
                        <p className="text-red-800 dark:text-red-200">{leverage.error}</p>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}
