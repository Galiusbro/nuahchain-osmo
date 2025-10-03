import React, { useState, useEffect } from 'react';
import Decimal from 'decimal.js';
import {
    Position,
    PositionSide,
    PositionStatus,
    positionSideToString,
    positionStatusToString,
    QueryPositionsByTraderResponse,
    QueryEstimatePositionResponse,
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
            
            await onOpenPosition({
                tokenDenom,
                collateralAmount: collateralAmountMicro,
                collateralDenom,
                leverage,
                side,
                minPrice: '0', // For simplicity, no slippage protection in this demo
                maxPrice: new Decimal(tokenPrice).mul(1.1).toString() // 10% slippage tolerance
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
                '0', // Min price (no slippage protection for demo)
                new Decimal(tokenPrice).mul(1.1).toString() // Max price with 10% slippage
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
        <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
            <h2>🚀 Leverage Trading (Max 100x)</h2>
            
            {error && (
                <div style={{ 
                    color: 'red', 
                    backgroundColor: '#ffebee', 
                    padding: '10px', 
                    borderRadius: '4px', 
                    marginBottom: '20px' 
                }}>
                    {error}
                </div>
            )}

            {/* Open Position Form */}
            <div style={{ 
                backgroundColor: '#f5f5f5', 
                padding: '20px', 
                borderRadius: '8px', 
                marginBottom: '30px' 
            }}>
                <h3>📈 Open New Position</h3>
                <form onSubmit={handleOpenPosition}>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '15px' }}>
                        <div>
                            <label>Token Denom:</label>
                            <input
                                type="text"
                                value={tokenDenom}
                                onChange={(e) => setTokenDenom(e.target.value)}
                                placeholder="factory/osmo.../mytoken"
                                style={{ width: '100%', padding: '8px', marginTop: '5px' }}
                                required
                            />
                        </div>
                        
                        <div>
                            <label>Collateral Amount:</label>
                            <input
                                type="number"
                                value={collateralAmount}
                                onChange={(e) => setCollateralAmount(e.target.value)}
                                placeholder="100"
                                min="0"
                                step="0.000001"
                                style={{ width: '100%', padding: '8px', marginTop: '5px' }}
                                required
                            />
                        </div>
                        
                        <div>
                            <label>Collateral Denom:</label>
                            <select
                                value={collateralDenom}
                                onChange={(e) => setCollateralDenom(e.target.value)}
                                style={{ width: '100%', padding: '8px', marginTop: '5px' }}
                            >
                                <option value="unuah">UNUAH</option>
                                <option value="factory/osmo1.../ndollar">NDollar</option>
                            </select>
                        </div>
                        
                        <div>
                            <label>Leverage (1x - 100x):</label>
                            <input
                                type="number"
                                value={leverage}
                                onChange={(e) => setLeverage(e.target.value)}
                                min="1"
                                max="100"
                                step="0.1"
                                style={{ width: '100%', padding: '8px', marginTop: '5px' }}
                                required
                            />
                        </div>
                        
                        <div>
                            <label>Position Side:</label>
                            <select
                                value={side}
                                onChange={(e) => setSide(parseInt(e.target.value) as PositionSide)}
                                style={{ width: '100%', padding: '8px', marginTop: '5px' }}
                            >
                                <option value={PositionSide.LONG}>LONG (Buy)</option>
                                <option value={PositionSide.SHORT}>SHORT (Sell)</option>
                            </select>
                        </div>
                    </div>
                    
                    {/* Position Estimation */}
                    {estimation && (
                        <div style={{ 
                            marginTop: '15px', 
                            padding: '10px', 
                            backgroundColor: '#e3f2fd', 
                            borderRadius: '4px' 
                        }}>
                            <h4>📊 Position Estimate:</h4>
                            <p><strong>Position Size:</strong> {formatAmount(estimation.positionSize)} tokens</p>
                            <p><strong>Entry Price:</strong> {formatPrice(estimation.entryPrice)} {collateralDenom}</p>
                            <p><strong>Liquidation Price:</strong> {formatPrice(estimation.liquidationPrice)} {collateralDenom}</p>
                            <p><strong>Trading Fee:</strong> {formatAmount(estimation.tradingFee)} {collateralDenom}</p>
                        </div>
                    )}
                    
                    <button
                        type="submit"
                        disabled={loading || !walletAddress}
                        style={{
                            marginTop: '15px',
                            padding: '12px 24px',
                            backgroundColor: side === PositionSide.LONG ? '#4caf50' : '#f44336',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: loading ? 'not-allowed' : 'pointer',
                            fontSize: '16px'
                        }}
                    >
                        {loading ? 'Opening...' : `Open ${positionSideToString(side)} Position`}
                    </button>
                </form>
            </div>

            {/* Current Positions */}
            <div>
                <h3>📋 Your Positions</h3>
                {loading && positions.length === 0 ? (
                    <p>Loading positions...</p>
                ) : positions.length === 0 ? (
                    <p>No positions found. Open your first leverage position above!</p>
                ) : (
                    <div style={{ overflowX: 'auto' }}>
                        <table style={{ width: '100%', borderCollapse: 'collapse', marginTop: '10px' }}>
                            <thead>
                                <tr style={{ backgroundColor: '#f0f0f0' }}>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>ID</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>Token</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>Side</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>Size</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>Collateral</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>Leverage</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>Entry Price</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>Liq. Price</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>PnL</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>Status</th>
                                    <th style={{ padding: '12px', textAlign: 'left', border: '1px solid #ddd' }}>Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {positions.map((position) => (
                                    <tr key={position.id}>
                                        <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                                            {position.id.substring(0, 8)}...
                                        </td>
                                        <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                                            {position.tokenDenom.split('/').pop()}
                                        </td>
                                        <td style={{ 
                                            padding: '12px', 
                                            border: '1px solid #ddd',
                                            color: position.side === PositionSide.LONG ? '#4caf50' : '#f44336',
                                            fontWeight: 'bold'
                                        }}>
                                            {positionSideToString(position.side)}
                                        </td>
                                        <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                                            {formatAmount(position.size)}
                                        </td>
                                        <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                                            {formatAmount(position.collateral)} {position.collateralDenom}
                                        </td>
                                        <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                                            {new Decimal(position.leverage).toFixed(1)}x
                                        </td>
                                        <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                                            {formatPrice(position.entryPrice)}
                                        </td>
                                        <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                                            {formatPrice(position.liquidationPrice)}
                                        </td>
                                        <td style={{ 
                                            padding: '12px', 
                                            border: '1px solid #ddd',
                                            color: new Decimal(position.unrealizedPnl).isPositive() ? '#4caf50' : '#f44336'
                                        }}>
                                            {new Decimal(position.unrealizedPnl).isPositive() ? '+' : ''}
                                            {formatAmount(position.unrealizedPnl)}
                                        </td>
                                        <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                                            <span style={{
                                                padding: '4px 8px',
                                                borderRadius: '4px',
                                                fontSize: '12px',
                                                backgroundColor: 
                                                    position.status === PositionStatus.OPEN ? '#e8f5e8' :
                                                    position.status === PositionStatus.CLOSED ? '#f0f0f0' : '#ffebee',
                                                color:
                                                    position.status === PositionStatus.OPEN ? '#2e7d32' :
                                                    position.status === PositionStatus.CLOSED ? '#666' : '#c62828'
                                            }}>
                                                {positionStatusToString(position.status)}
                                            </span>
                                        </td>
                                        <td style={{ padding: '12px', border: '1px solid #ddd' }}>
                                            {position.status === PositionStatus.OPEN && (
                                                <button
                                                    onClick={() => handleClosePosition(position.id)}
                                                    disabled={loading}
                                                    style={{
                                                        padding: '6px 12px',
                                                        backgroundColor: '#ff9800',
                                                        color: 'white',
                                                        border: 'none',
                                                        borderRadius: '4px',
                                                        cursor: loading ? 'not-allowed' : 'pointer',
                                                        fontSize: '12px'
                                                    }}
                                                >
                                                    Close
                                                </button>
                                            )}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    );
};

export default LeverageTrading;
