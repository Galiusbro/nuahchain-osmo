import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import type { TokenFormState } from '@/types';
import { Plus } from 'lucide-react';
import type { ChangeEvent, FormEvent } from "react";

interface CreateTokenFormProps {
    form: TokenFormState;
    onInputChange: (field: keyof TokenFormState) => (event: ChangeEvent<HTMLInputElement>) => void;
    onSubmit: (event: FormEvent<HTMLFormElement>) => void;
    isSubmitting: boolean;
    walletAddress: string;
}

export default function CreateTokenForm({
    form,
    onInputChange,
    onSubmit,
    isSubmitting,
    walletAddress
}: CreateTokenFormProps) {
    if (!walletAddress) {
        return (
            <Card className="mb-8">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Plus className="h-5 w-5" />
                        Create Token
                    </CardTitle>
                    <CardDescription>
                        Connect your wallet to create a new token
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <p className="text-muted-foreground text-center py-8">
                        Please connect your wallet first to create tokens.
                    </p>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className="mb-8">
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <Plus className="h-5 w-5" />
                    Create Token
                </CardTitle>
                <CardDescription>
                    Create a new user token on the blockchain
                </CardDescription>
            </CardHeader>
            <CardContent>
                <form onSubmit={onSubmit} className="space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="subdenom">Subdenom *</Label>
                            <Input
                                id="subdenom"
                                value={form.subdenom}
                                onChange={onInputChange("subdenom")}
                                placeholder="mytoken"
                                required
                            />
                            <p className="text-sm text-muted-foreground">
                                Unique identifier for your token (lowercase, alphanumeric)
                            </p>
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="name">Name *</Label>
                            <Input
                                id="name"
                                value={form.name}
                                onChange={onInputChange("name")}
                                placeholder="My Token"
                                required
                            />
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="symbol">Symbol *</Label>
                            <Input
                                id="symbol"
                                value={form.symbol}
                                onChange={onInputChange("symbol")}
                                placeholder="MTK"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="decimals">Decimals</Label>
                            <Input
                                id="decimals"
                                type="number"
                                min={0}
                                max={18}
                                value={form.decimals}
                                onChange={onInputChange("decimals")}
                                required
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="memo">Memo (optional)</Label>
                        <Input
                            id="memo"
                            value={form.memo}
                            onChange={onInputChange("memo")}
                            placeholder="Optional transaction memo"
                        />
                    </div>

                    <div className="flex justify-center pt-4">
                        <Button
                            type="submit"
                            disabled={isSubmitting}
                            className="w-full md:w-auto"
                        >
                            {isSubmitting ? "Creating Token..." : "Create Token"}
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}




