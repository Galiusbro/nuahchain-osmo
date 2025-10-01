import React, { useEffect, useState, useCallback } from 'react';
import { SigningStargateClient } from '@cosmjs/stargate';

interface Balance {
    denom: string;
    amount: string;
}

interface WalletBalanceProps {
    client: SigningStargateClient | null;
    walletAddress: string;
    restEndpoint: string;
    onError?: (error: string) => void;
}

const WalletBalance: React.FC<WalletBalanceProps> = ({
    client,
    walletAddress,
    restEndpoint,
    onError
}) => {
    const [balances, setBalances] = useState<Balance[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const formatAmount = (amount: string, decimals: number = 6): string => {
        const num = parseFloat(amount);
        if (isNaN(num)) return '0';
        
        const divisor = Math.pow(10, decimals);
        const formatted = (num / divisor).toFixed(decimals);
        
        // Remove trailing zeros
        return parseFloat(formatted).toString();
    };

    const formatDenom = (denom: string): string => {
        // Handle factory tokens
        if (denom.startsWith('factory/')) {
            const parts = denom.split('/');
            return parts[parts.length - 1] || denom;
        }
        
        // Handle IBC tokens
        if (denom.startsWith('ibc/')) {
            return `IBC/${denom.slice(4, 12)}...`;
        }
        
        // Handle native tokens
        if (denom.startsWith('u')) {
            return denom.slice(1).toUpperCase();
        }
        
        return denom.toUpperCase();
    };

    const fetchBalances = useCallback(async () => {
        if (!walletAddress || !restEndpoint) {
            setBalances([]);
            return;
        }

        setIsLoading(true);
        setError(null);

        try {
            const baseUrl = restEndpoint.trim().replace(/\/+$/, '');
            const response = await fetch(`${baseUrl}/cosmos/bank/v1beta1/balances/${walletAddress}`);
            
            if (!response.ok) {
                throw new Error(`Failed to fetch balances: ${response.status}`);
            }

            const data = await response.json();
            const fetchedBalances = data.balances || [];
            
            // Filter out zero balances and sort by amount (descending)
            const nonZeroBalances = fetchedBalances
                .filter((balance: Balance) => balance.amount !== '0')
                .sort((a: Balance, b: Balance) => {
                    const amountA = parseFloat(a.amount);
                    const amountB = parseFloat(b.amount);
                    return amountB - amountA;
                });

            setBalances(nonZeroBalances);
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Failed to fetch balances';
            setError(errorMessage);
            onError?.(errorMessage);
        } finally {
            setIsLoading(false);
        }
    }, [walletAddress, restEndpoint, onError]);

    useEffect(() => {
        fetchBalances();
    }, [fetchBalances]);

    const containerStyle: React.CSSProperties = {
        backgroundColor: '#f8fafc',
        border: '1px solid #e2e8f0',
        borderRadius: '8px',
        padding: '16px',
        marginBottom: '20px'
    };

    const headerStyle: React.CSSProperties = {
        fontSize: '18px',
        fontWeight: 600,
        marginBottom: '12px',
        color: '#2d3748',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between'
    };

    const refreshButtonStyle: React.CSSProperties = {
        backgroundColor: '#4299e1',
        color: 'white',
        border: 'none',
        borderRadius: '4px',
        padding: '4px 8px',
        fontSize: '12px',
        cursor: 'pointer',
        fontWeight: 500
    };

    const balanceItemStyle: React.CSSProperties = {
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '8px 0',
        borderBottom: '1px solid #e2e8f0'
    };

    const denomStyle: React.CSSProperties = {
        fontWeight: 500,
        color: '#4a5568'
    };

    const amountStyle: React.CSSProperties = {
        fontWeight: 600,
        color: '#2d3748'
    };

    const errorStyle: React.CSSProperties = {
        color: '#e53e3e',
        fontSize: '14px',
        fontStyle: 'italic'
    };

    const emptyStyle: React.CSSProperties = {
        color: '#718096',
        fontSize: '14px',
        fontStyle: 'italic',
        textAlign: 'center',
        padding: '20px 0'
    };

    if (!walletAddress) {
        return (
            <div style={containerStyle}>
                <div style={headerStyle}>Wallet Balances</div>
                <div style={emptyStyle}>Connect wallet to view balances</div>
            </div>
        );
    }

    return (
        <div style={containerStyle}>
            <div style={headerStyle}>
                Wallet Balances
                <button 
                    style={refreshButtonStyle}
                    onClick={fetchBalances}
                    disabled={isLoading}
                >
                    {isLoading ? 'Loading...' : 'Refresh'}
                </button>
            </div>
            
            {error && (
                <div style={errorStyle}>
                    Error: {error}
                </div>
            )}
            
            {!error && balances.length === 0 && !isLoading && (
                <div style={emptyStyle}>
                    No balances found
                </div>
            )}
            
            {!error && balances.length > 0 && (
                <div>
                    {balances.map((balance, index) => (
                        <div key={`${balance.denom}-${index}`} style={balanceItemStyle}>
                            <span style={denomStyle}>
                                {formatDenom(balance.denom)}
                            </span>
                            <span style={amountStyle}>
                                {formatAmount(balance.amount)}
                            </span>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default WalletBalance;