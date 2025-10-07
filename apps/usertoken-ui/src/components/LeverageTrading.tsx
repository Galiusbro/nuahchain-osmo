import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import Decimal from 'decimal.js';
import { RefreshCw, TrendingDown, TrendingUp } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import {
    Position,
    PositionSide,
    positionSideToString,
    PositionStatus,
    positionStatusToString,
    QueryEstimatePositionResponse,
    QueryPositionsByTraderResponse,
    QueryTokenPriceResponse
} from '../codec/leverage';

interface LeverageTradingProps {
    walletAddress: string;
    restEndpoint: string;
    onOpenPosition: (params: OpenPositionParams) => Promise<void>;
    onClosePosition: (positionId: string, minPrice: string, maxPrice: string) => Promise<void>;
    onAddCollateral: (positionId: string, amount: string, denom: string) => Promise<void>;
}

interface OpenPositionParams {
    tokenDenom: string;
    collateralAmount: string;
    collateralDenom: string;
    leverage: string;
    side: PositionSide;
    minPrice: string;
    maxPrice: string;
}

const LeverageTrading: React.FC<LeverageTradingProps> = ({
    walletAddress,
    restEndpoint,
    onOpenPosition,
    onClosePosition,
    onAddCollateral
}) => {
    const [positions, setPositions] = useState<Position[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string>('');

    // Form state for opening positions
    const [tokenDenom, setTokenDenom] = useState('');
    const [collateralAmount, setCollateralAmount] = useState('');
    const [collateralDenom, setCollateralDenom] = useState('unuah');
    const [leverage, setLeverage] = useState('2');
    const [side, setSide] = useState<PositionSide>(PositionSide.LONG);

    // Price and estimation data
    const [tokenPrice, setTokenPrice] = useState<string>('0');
    const [estimation, setEstimation] = useState<QueryEstimatePositionResponse | null>(null);

    // Load user positions
    const loadPositions = async () => {
        if (!walletAddress) return;

        setLoading(true);
        try {
            const response = await fetch(
                `${restEndpoint}/osmosis/leverage/v1beta1/positions/trader/${walletAddress}`
            );

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data: QueryPositionsByTraderResponse = await response.json();
            setPositions(data.positions || []);
        } catch (err) {
            console.error('Failed to load positions:', err);
            setError(err instanceof Error ? err.message : 'Failed to load positions');
        } finally {
            setLoading(false);
        }
    };

    // Load token price
    const loadTokenPrice = async (denom: string) => {
        if (!denom) return;

        try {
            const response = await fetch(
                `${restEndpoint}/osmosis/leverage/v1beta1/token-price/${encodeURIComponent(denom)}`
            );

            if (response.ok) {
                const data: QueryTokenPriceResponse = await response.json();
                setTokenPrice(data.price);
            }
        } catch (err) {
            console.error('Failed to load token price:', err);
        }
    };

    // Estimate position
    const estimatePosition = async () => {
        if (!tokenDenom || !collateralAmount || !leverage) return;

        try {
            const params = new URLSearchParams({
                tokenDenom,
                collateralAmount: new Decimal(collateralAmount).mul(1_000_000).toString(), // Convert to micro units
                leverage,
                side: side.toString()
            });

            const response = await fetch(
                `${restEndpoint}/osmosis/leverage/v1beta1/estimate-position?${params}`
            );

            if (response.ok) {
                const data: QueryEstimatePositionResponse = await response.json();
                setEstimation(data);
            }
        } catch (err) {
            console.error('Failed to estimate position:', err);
        }
    };

    // Load data on component mount and when dependencies change
    useEffect(() => {
        loadPositions();
    }, [walletAddress]);

    useEffect(() => {
        loadTokenPrice(tokenDenom);
    }, [tokenDenom]);

    useEffect(() => {
        estimatePosition();
    }, [tokenDenom, collateralAmount, leverage, side]);

    // Handle form submission
    const handleOpenPosition = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!tokenDenom || !collateralAmount || !leverage) {
            setError('Please fill in all required fields');
            return;
        }

        try {
            setLoading(true);
            setError('');

            const collateralAmountMicro = new Decimal(collateralAmount).mul(1_000_000).toString();

            const priceMicro = tokenPrice && tokenPrice !== '0'
                ? new Decimal(tokenPrice).mul(1_000_000)
                : null;

            await onOpenPosition({
                tokenDenom,
                collateralAmount: collateralAmountMicro,
                collateralDenom,
                leverage,
                side,
                minPrice: side === PositionSide.SHORT && priceMicro
                    ? priceMicro.mul(0.9).toFixed(18)
                    : '0',
                maxPrice: side === PositionSide.LONG && priceMicro
                    ? priceMicro.mul(1.1).toFixed(18)
                    : '999999999000000'
            });

            // Reload positions after successful trade
            await loadPositions();

            // Reset form
            setTokenDenom('');
            setCollateralAmount('');
            setLeverage('2');
        } catch (err) {
            console.error('Failed to open position:', err);
            setError(err instanceof Error ? err.message : 'Failed to open position');
        } finally {
            setLoading(false);
        }
    };

    // Handle position closing
    const handleClosePosition = async (positionId: string) => {
        try {
            setLoading(true);
            setError('');

            await onClosePosition(
                positionId,
                '0', // Keep wide slippage bounds when closing positions
                '999999999000000'
            );

            // Reload positions after successful close
            await loadPositions();
        } catch (err) {
            console.error('Failed to close position:', err);
            setError(err instanceof Error ? err.message : 'Failed to close position');
        } finally {
            setLoading(false);
        }
    };

    // Format numbers for display
    const formatAmount = (amount: string, decimals: number = 6): string => {
        return new Decimal(amount).div(Math.pow(10, decimals)).toFixed(decimals);
    };

    const formatPrice = (price: string): string => {
        return new Decimal(price).toFixed(6);
    };

    return (
        <div className="p-5 max-w-6xl mx-auto">
            <div className="mb-6">
                <h2 className="text-3xl font-bold flex items-center gap-2">
                    🚀 Leverage Trading (Max 100x)
                </h2>
            </div>

            {error && (
                <div className="bg-destructive/10 text-destructive p-3 rounded-md mb-5">
                    {error}
                </div>
            )}

            <Tabs defaultValue="open" className="space-y-6">
                <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="open">Open Position</TabsTrigger>
                    <TabsTrigger value="positions">Your Positions</TabsTrigger>
                </TabsList>

                <TabsContent value="open">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <TrendingUp className="h-5 w-5" />
                                Open New Position
                            </CardTitle>
                            <CardDescription>
                                Create a new leverage position with up to 100x leverage
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <form onSubmit={handleOpenPosition} className="space-y-6">
                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="tokenDenom">Token Denom</Label>
                                        <Input
                                            id="tokenDenom"
                                            type="text"
                                            value={tokenDenom}
                                            onChange={(e) => setTokenDenom(e.target.value)}
                                            placeholder="factory/osmo.../mytoken"
                                            required
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="collateralAmount">Collateral Amount</Label>
                                        <Input
                                            id="collateralAmount"
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
                                        <Label htmlFor="collateralDenom">Collateral Denom</Label>
                                        <select
                                            id="collateralDenom"
                                            value={collateralDenom}
                                            onChange={(e) => setCollateralDenom(e.target.value)}
                                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                        >
                                            <option value="unuah">UNUAH</option>
                                            <option value="factory/osmo1.../ndollar">NDollar</option>
                                        </select>
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="leverage">Leverage (1x - 100x)</Label>
                                        <Input
                                            id="leverage"
                                            type="number"
                                            value={leverage}
                                            onChange={(e) => setLeverage(e.target.value)}
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
                                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                        >
                                            <option value={PositionSide.LONG}>LONG (Buy)</option>
                                            <option value={PositionSide.SHORT}>SHORT (Sell)</option>
                                        </select>
                                    </div>
                                </div>

                                {/* Position Estimation */}
                                {estimation && (
                                    <Card className="bg-blue-50 dark:bg-blue-950/20">
                                        <CardHeader>
                                            <CardTitle className="text-lg">📊 Position Estimate</CardTitle>
                                        </CardHeader>
                                        <CardContent className="space-y-2">
                                            <div className="flex justify-between">
                                                <span className="font-medium">Position Size:</span>
                                                <span>{formatAmount(estimation.positionSize)} tokens</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span className="font-medium">Entry Price:</span>
                                                <span>{formatPrice(estimation.entryPrice)} {collateralDenom}</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span className="font-medium">Liquidation Price:</span>
                                                <span>{formatPrice(estimation.liquidationPrice)} {collateralDenom}</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span className="font-medium">Trading Fee:</span>
                                                <span>{formatAmount(estimation.tradingFee)} {collateralDenom}</span>
                                            </div>
                                        </CardContent>
                                    </Card>
                                )}

                                <Button
                                    type="submit"
                                    disabled={loading || !walletAddress}
                                    className={`w-full ${side === PositionSide.LONG
                                            ? 'bg-green-600 hover:bg-green-700'
                                            : 'bg-red-600 hover:bg-red-700'
                                        }`}
                                >
                                    {loading ? (
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

                <TabsContent value="positions">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                📋 Your Positions
                            </CardTitle>
                            <CardDescription>
                                Manage your active leverage positions
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {loading && positions.length === 0 ? (
                                <div className="flex items-center justify-center py-8">
                                    <RefreshCw className="h-6 w-6 animate-spin mr-2" />
                                    <span>Loading positions...</span>
                                </div>
                            ) : positions.length === 0 ? (
                                <div className="text-center py-8">
                                    <p className="text-muted-foreground">
                                        No positions found. Open your first leverage position above!
                                    </p>
                                </div>
                            ) : (
                                <div className="overflow-x-auto">
                                    <table className="w-full border-collapse">
                                        <thead>
                                            <tr className="border-b">
                                                <th className="text-left p-3 font-medium">ID</th>
                                                <th className="text-left p-3 font-medium">Token</th>
                                                <th className="text-left p-3 font-medium">Side</th>
                                                <th className="text-left p-3 font-medium">Size</th>
                                                <th className="text-left p-3 font-medium">Collateral</th>
                                                <th className="text-left p-3 font-medium">Leverage</th>
                                                <th className="text-left p-3 font-medium">Entry Price</th>
                                                <th className="text-left p-3 font-medium">Liq. Price</th>
                                                <th className="text-left p-3 font-medium">PnL</th>
                                                <th className="text-left p-3 font-medium">Status</th>
                                                <th className="text-left p-3 font-medium">Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {positions.map((position) => (
                                                <tr key={position.id} className="border-b hover:bg-muted/50">
                                                    <td className="p-3 text-sm font-mono">
                                                        {position.id.substring(0, 8)}...
                                                    </td>
                                                    <td className="p-3">
                                                        {position.tokenDenom.split('/').pop()}
                                                    </td>
                                                    <td className={`p-3 font-bold ${position.side === PositionSide.LONG
                                                            ? 'text-green-600'
                                                            : 'text-red-600'
                                                        }`}>
                                                        {positionSideToString(position.side)}
                                                    </td>
                                                    <td className="p-3">
                                                        {formatAmount(position.size)}
                                                    </td>
                                                    <td className="p-3">
                                                        {formatAmount(position.collateral)} {position.collateralDenom}
                                                    </td>
                                                    <td className="p-3">
                                                        {new Decimal(position.leverage).toFixed(1)}x
                                                    </td>
                                                    <td className="p-3">
                                                        {formatPrice(position.entryPrice)}
                                                    </td>
                                                    <td className="p-3">
                                                        {formatPrice(position.liquidationPrice)}
                                                    </td>
                                                    <td className={`p-3 font-medium ${new Decimal(position.unrealizedPnl).isPositive()
                                                            ? 'text-green-600'
                                                            : 'text-red-600'
                                                        }`}>
                                                        {new Decimal(position.unrealizedPnl).isPositive() ? '+' : ''}
                                                        {formatAmount(position.unrealizedPnl)}
                                                    </td>
                                                    <td className="p-3">
                                                        <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${position.status === PositionStatus.OPEN
                                                                ? 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400'
                                                                : position.status === PositionStatus.CLOSED
                                                                    ? 'bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400'
                                                                    : 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400'
                                                            }`}>
                                                            {positionStatusToString(position.status)}
                                                        </span>
                                                    </td>
                                                    <td className="p-3">
                                                        {position.status === PositionStatus.OPEN && (
                                                            <Button
                                                                onClick={() => handleClosePosition(position.id)}
                                                                disabled={loading}
                                                                size="sm"
                                                                variant="destructive"
                                                            >
                                                                Close
                                                            </Button>
                                                        )}
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
};

export default LeverageTrading;
