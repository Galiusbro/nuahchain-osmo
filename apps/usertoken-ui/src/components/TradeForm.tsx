import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import type { TradeFormState, TradeMode, UserToken } from '@/types';
import { Coins } from 'lucide-react';
import type { ChangeEvent, FormEvent } from "react";

interface TradeFormProps {
    tradeForm: TradeFormState;
    tradeMode: TradeMode;
    onTradeModeChange: (mode: TradeMode) => void;
    onInputChange: (field: keyof TradeFormState) => (event: ChangeEvent<HTMLInputElement>) => void;
    onSubmit: (event: FormEvent<HTMLFormElement>) => void;
    isSubmitting: boolean;
    walletAddress: string;
    availableBuyTokens: UserToken[];
    availableSellTokens: UserToken[];
    tradePreview?: {
        tokensUnits?: any;
        payoutNUAH?: any;
        loading?: boolean;
        error?: string;
    };
}

export default function TradeForm({
    tradeForm,
    tradeMode,
    onTradeModeChange,
    onInputChange,
    onSubmit,
    isSubmitting,
    walletAddress,
    availableBuyTokens,
    availableSellTokens,
    tradePreview
}: TradeFormProps) {
    if (!walletAddress) {
        return (
            <Card className="mb-8">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Coins className="h-5 w-5" />
                        Trade Tokens
                    </CardTitle>
                    <CardDescription>
                        Connect your wallet to trade tokens
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <p className="text-muted-foreground text-center py-8">
                        Please connect your wallet first to trade tokens.
                    </p>
                </CardContent>
            </Card>
        );
    }

    const availableTokens = tradeMode === "buy" ? availableBuyTokens : availableSellTokens;

    return (
        <Card className="mb-8">
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <Coins className="h-5 w-5" />
                    Trade Tokens
                </CardTitle>
                <CardDescription>
                    Buy or sell tokens on the bonding curve
                </CardDescription>
            </CardHeader>
            <CardContent>
                <div className="flex gap-2 mb-6">
                    <Button
                        type="button"
                        variant={tradeMode === "buy" ? "default" : "outline"}
                        onClick={() => onTradeModeChange("buy")}
                        className="flex-1"
                    >
                        Buy Tokens
                    </Button>
                    <Button
                        type="button"
                        variant={tradeMode === "sell" ? "default" : "outline"}
                        onClick={() => onTradeModeChange("sell")}
                        className="flex-1"
                    >
                        Sell Tokens
                    </Button>
                </div>

                <form onSubmit={onSubmit} className="space-y-6">
                    <div className="space-y-2">
                        <Label htmlFor="token-denom">Select Token</Label>
                        <select
                            id="token-denom"
                            value={tradeForm.denom}
                            onChange={(e) => onInputChange("denom")({ target: { value: e.target.value } } as any)}
                            className="w-full p-2 border border-gray-300 rounded-md"
                            required
                        >
                            <option value="">Choose a token...</option>
                            {availableTokens.map((token) => (
                                <option key={token.denom} value={token.denom}>
                                    {token.symbol || token.name || token.denom}
                                </option>
                            ))}
                        </select>
                    </div>

                    {tradeMode === "buy" ? (
                        <>
                            <div className="space-y-2">
                                <Label htmlFor="payment-amount">Payment Amount (NUAH)</Label>
                                <Input
                                    id="payment-amount"
                                    type="number"
                                    step="0.000001"
                                    min="0"
                                    value={tradeForm.paymentAmount}
                                    onChange={onInputChange("paymentAmount")}
                                    placeholder="Enter amount to spend"
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="min-tokens">Minimum Tokens (optional)</Label>
                                <Input
                                    id="min-tokens"
                                    type="number"
                                    step="0.000001"
                                    min="0"
                                    value={tradeForm.minTokens}
                                    onChange={onInputChange("minTokens")}
                                    placeholder="Minimum tokens to receive"
                                />
                            </div>
                            {tradePreview?.tokensUnits && (
                                <div className="p-4 bg-blue-50 border border-blue-200 rounded-md">
                                    <p className="text-sm text-blue-800">
                                        <strong>Estimated tokens:</strong> {tradePreview.tokensUnits.toString()}
                                    </p>
                                </div>
                            )}
                        </>
                    ) : (
                        <>
                            <div className="space-y-2">
                                <Label htmlFor="token-amount">Token Amount</Label>
                                <Input
                                    id="token-amount"
                                    type="number"
                                    step="0.000001"
                                    min="0"
                                    value={tradeForm.tokenAmount}
                                    onChange={onInputChange("tokenAmount")}
                                    placeholder="Enter amount to sell"
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="min-price">Minimum Price (NUAH)</Label>
                                <Input
                                    id="min-price"
                                    type="number"
                                    step="0.000001"
                                    min="0"
                                    value={tradeForm.minPrice}
                                    onChange={onInputChange("minPrice")}
                                    placeholder="Minimum price per token"
                                />
                            </div>
                            {tradePreview?.payoutNUAH && (
                                <div className="p-4 bg-green-50 border border-green-200 rounded-md">
                                    <p className="text-sm text-green-800">
                                        <strong>Estimated payout:</strong> {tradePreview.payoutNUAH.toString()} NUAH
                                    </p>
                                </div>
                            )}
                        </>
                    )}

                    {tradePreview?.error && (
                        <div className="p-4 bg-red-50 border border-red-200 rounded-md">
                            <p className="text-sm text-red-800">{tradePreview.error}</p>
                        </div>
                    )}

                    <div className="flex justify-center pt-4">
                        <Button
                            type="submit"
                            disabled={isSubmitting || availableTokens.length === 0}
                            className="w-full md:w-auto"
                        >
                            {isSubmitting
                                ? `${tradeMode === "buy" ? "Buying" : "Selling"}...`
                                : `${tradeMode === "buy" ? "Buy" : "Sell"} Tokens`
                            }
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}


