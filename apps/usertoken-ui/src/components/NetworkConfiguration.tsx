import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import type { NetworkConfigState } from '@/types';
import { Settings, Wallet } from 'lucide-react';
import type { ChangeEvent } from "react";

interface NetworkConfigurationProps {
    network: NetworkConfigState;
    onNetworkChange: (field: keyof NetworkConfigState) => (event: ChangeEvent<HTMLInputElement>) => void;
    onConnectWallet: () => void;
    isConnecting: boolean;
    walletAddress: string;
    status?: string | null;
    error?: string | null;
}

export default function NetworkConfiguration({
    network,
    onNetworkChange,
    onConnectWallet,
    isConnecting,
    walletAddress,
    status,
    error
}: NetworkConfigurationProps) {
    return (
        <>
            <Card className="mb-8">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Settings className="h-5 w-5" />
                        Network Configuration
                    </CardTitle>
                    <CardDescription>
                        Configure your blockchain network settings
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="chain-id">Chain ID</Label>
                            <Input
                                id="chain-id"
                                value={network.chainId}
                                onChange={onNetworkChange("chainId")}
                                placeholder="nuahchain"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="chain-name">Chain name</Label>
                            <Input
                                id="chain-name"
                                value={network.chainName}
                                onChange={onNetworkChange("chainName")}
                                placeholder="Nuahchain"
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="bech32-prefix">Bech32 prefix</Label>
                            <Input
                                id="bech32-prefix"
                                value={network.bech32Prefix}
                                onChange={onNetworkChange("bech32Prefix")}
                                placeholder="nuah"
                                required
                            />
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="rpc-endpoint">RPC endpoint</Label>
                            <Input
                                id="rpc-endpoint"
                                value={network.rpcEndpoint}
                                onChange={onNetworkChange("rpcEndpoint")}
                                placeholder="http://localhost:5173/api"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="rest-endpoint">REST endpoint</Label>
                            <Input
                                id="rest-endpoint"
                                value={network.restEndpoint}
                                onChange={onNetworkChange("restEndpoint")}
                                placeholder="http://localhost:5173/rest"
                                required
                            />
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="coin-denom">Display denom</Label>
                            <Input
                                id="coin-denom"
                                value={network.coinDenom}
                                onChange={onNetworkChange("coinDenom")}
                                placeholder="NUAH"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="coin-minimal-denom">Minimal denom</Label>
                            <Input
                                id="coin-minimal-denom"
                                value={network.coinMinimalDenom}
                                onChange={onNetworkChange("coinMinimalDenom")}
                                placeholder="unuah"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="coin-decimals">Decimals</Label>
                            <Input
                                id="coin-decimals"
                                type="number"
                                min={0}
                                max={18}
                                value={network.coinDecimals}
                                onChange={onNetworkChange("coinDecimals")}
                                required
                            />
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="gas-price-low">Gas price (low)</Label>
                            <Input
                                id="gas-price-low"
                                value={network.gasPriceLow}
                                onChange={onNetworkChange("gasPriceLow")}
                                placeholder="0.005"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="gas-price-average">Gas price (average)</Label>
                            <Input
                                id="gas-price-average"
                                value={network.gasPriceAverage}
                                onChange={onNetworkChange("gasPriceAverage")}
                                placeholder="0.025"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="gas-price-high">Gas price (high)</Label>
                            <Input
                                id="gas-price-high"
                                value={network.gasPriceHigh}
                                onChange={onNetworkChange("gasPriceHigh")}
                                placeholder="0.04"
                                required
                            />
                        </div>
                    </div>

                    <div className="flex justify-center pt-4">
                        <Button
                            onClick={onConnectWallet}
                            disabled={isConnecting}
                            className="w-full md:w-auto"
                        >
                            {isConnecting ? "Connecting..." : walletAddress ? "Reconnect Wallet" : "Connect Keplr"}
                        </Button>
                    </div>
                </CardContent>
            </Card>

            {/* Wallet Status */}
            {walletAddress && (
                <Card className="mb-6">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Wallet className="h-5 w-5" />
                            Wallet Connected
                        </CardTitle>
                        <CardDescription>
                            Address: {walletAddress}
                        </CardDescription>
                    </CardHeader>
                </Card>
            )}

            {/* Status Messages */}
            {status && (
                <Card className="mb-6 border-green-200 bg-green-50">
                    <CardContent className="pt-6">
                        <p className="text-green-800 font-medium">{status}</p>
                    </CardContent>
                </Card>
            )}

            {error && (
                <Card className="mb-6 border-red-200 bg-red-50">
                    <CardContent className="pt-6">
                        <p className="text-red-800 font-medium">{error}</p>
                    </CardContent>
                </Card>
            )}
        </>
    );
}




