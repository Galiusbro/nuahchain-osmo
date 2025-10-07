import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SigningStargateClient } from '@cosmjs/stargate';
import { RefreshCw } from 'lucide-react';
import React, { useCallback, useEffect, useState } from 'react';

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


    if (!walletAddress) {
        return (
            <Card className="mb-5">
                <CardHeader>
                    <CardTitle>Wallet Balances</CardTitle>
                </CardHeader>
                <CardContent>
                    <p className="text-muted-foreground text-center py-5">Connect wallet to view balances</p>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className="mb-5">
            <CardHeader>
                <div className="flex items-center justify-between">
                    <CardTitle>Wallet Balances</CardTitle>
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={fetchBalances}
                        disabled={isLoading}
                    >
                        <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                        {isLoading ? 'Loading...' : 'Refresh'}
                    </Button>
                </div>
            </CardHeader>

            <CardContent>
                {error && (
                    <div className="text-destructive text-sm italic mb-4">
                        Error: {error}
                    </div>
                )}

                {!error && balances.length === 0 && !isLoading && (
                    <p className="text-muted-foreground text-center py-5">
                        No balances found
                    </p>
                )}

                {!error && balances.length > 0 && (
                    <div className="space-y-2">
                        {balances.map((balance, index) => (
                            <div key={`${balance.denom}-${index}`} className="flex justify-between items-center py-2 border-b last:border-b-0">
                                <span className="font-medium text-muted-foreground">
                                    {formatDenom(balance.denom)}
                                </span>
                                <span className="font-semibold">
                                    {formatAmount(balance.amount)}
                                </span>
                            </div>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
    );
};

export default WalletBalance;
