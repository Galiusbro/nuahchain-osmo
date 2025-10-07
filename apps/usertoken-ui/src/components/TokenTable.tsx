import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import type { UserToken } from '@/types';
import { Eye } from 'lucide-react';
import type { ReactNode } from "react";

interface TokenTableProps {
    tokens: UserToken[];
    emptyMessage: string;
    isLoading?: boolean;
    error?: string | null;
    restBaseUrl?: string;
    renderActions?: (token: UserToken) => ReactNode;
}

const FOUNDER_TRANCHE_AMOUNT_DISPLAY = "10,000,000";

const formatAmount = (value?: string | null, decimals: number = 0) => {
    const raw = value && value.trim() !== "" ? value : "0";
    try {
        const big = BigInt(raw);

        if (decimals === 0) {
            return big.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
        }

        const divisor = BigInt(10 ** decimals);
        const wholePart = big / divisor;
        const fractionalPart = big % divisor;

        const wholeFormatted = wholePart.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");

        if (fractionalPart === 0n) {
            return wholeFormatted;
        }

        const fractionalStr = fractionalPart.toString().padStart(decimals, '0');
        const trimmedFractional = fractionalStr.replace(/0+$/, '');

        return trimmedFractional.length > 0 ? `${wholeFormatted}.${trimmedFractional}` : wholeFormatted;
    } catch {
        return raw;
    }
};

const isTokenWithProperDecimals = (token: UserToken): boolean => {
    try {
        const currentSupply = BigInt(token.current_supply || "0");
        const maxSupply = BigInt(token.max_supply || "0");
        return currentSupply > 1_000_000_000n || maxSupply > 1_000_000_000n;
    } catch {
        return false;
    }
};

const tableHeaderCellStyle: React.CSSProperties = {
    fontWeight: 600,
    textAlign: "left",
    padding: "10px 12px",
    backgroundColor: "#f7fafc",
    borderBottom: "1px solid #e2e8f0"
};

const tableCellStyle: React.CSSProperties = {
    padding: "10px 12px",
    borderBottom: "1px solid #e2e8f0"
};

export default function TokenTable({
    tokens,
    emptyMessage,
    isLoading = false,
    error,
    restBaseUrl,
    renderActions
}: TokenTableProps) {
    if (!restBaseUrl) {
        return (
            <Card>
                <CardContent className="pt-6">
                    <p className="text-muted-foreground text-center py-8">
                        Provide a REST endpoint to load token data.
                    </p>
                </CardContent>
            </Card>
        );
    }

    if (error) {
        return (
            <Card>
                <CardContent className="pt-6">
                    <p className="text-red-600 text-center py-8">{error}</p>
                </CardContent>
            </Card>
        );
    }

    if (isLoading) {
        return (
            <Card>
                <CardContent className="pt-6">
                    <p className="text-muted-foreground text-center py-8">Fetching tokens...</p>
                </CardContent>
            </Card>
        );
    }

    if (tokens.length === 0) {
        return (
            <Card>
                <CardContent className="pt-6">
                    <p className="text-muted-foreground text-center py-8">{emptyMessage}</p>
                </CardContent>
            </Card>
        );
    }

    const showActions = Boolean(renderActions);
    const filteredTokens = tokens.filter(token => isTokenWithProperDecimals(token));

    return (
        <Card>
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <Eye className="h-5 w-5" />
                    Tokens ({filteredTokens.length})
                </CardTitle>
                <CardDescription>
                    {showActions ? "Your created tokens" : "All available tokens"}
                </CardDescription>
            </CardHeader>
            <CardContent>
                <div style={{ overflowX: "auto" }}>
                    <table style={{ width: "100%", borderCollapse: "collapse" }}>
                        <thead>
                            <tr>
                                <th style={tableHeaderCellStyle}>Name</th>
                                <th style={tableHeaderCellStyle}>Symbol</th>
                                <th style={tableHeaderCellStyle}>Supply</th>
                                <th style={tableHeaderCellStyle}>Denom</th>
                                <th style={tableHeaderCellStyle}>Creator</th>
                                <th style={tableHeaderCellStyle}>Founder tranche</th>
                                {showActions && <th style={tableHeaderCellStyle}>Actions</th>}
                            </tr>
                        </thead>
                        <tbody>
                            {filteredTokens.map((token) => {
                                const founderClaimedRaw = token.founder_tokens_claimed ?? "0";
                                const founderTrancheClaimed = founderClaimedRaw !== "" && founderClaimedRaw !== "0";
                                const hasProperDecimals = isTokenWithProperDecimals(token);
                                const decimals = 6;

                                const founderStatus = founderTrancheClaimed
                                    ? `Claimed (${formatAmount(founderClaimedRaw, hasProperDecimals ? decimals : 0)})`
                                    : `Available (${FOUNDER_TRANCHE_AMOUNT_DISPLAY} reserved)`;

                                const currentSupply = formatAmount(token.current_supply, hasProperDecimals ? decimals : 0);
                                const maxSupply = formatAmount(token.max_supply, hasProperDecimals ? decimals : 0);

                                return (
                                    <tr key={token.denom}>
                                        <td style={tableCellStyle}>{token.name ?? "–"}</td>
                                        <td style={tableCellStyle}>
                                            <strong>{token.symbol ?? "–"}</strong>
                                        </td>
                                        <td style={tableCellStyle}>
                                            <div style={{ fontSize: "12px" }}>
                                                <div><strong>Current:</strong> {currentSupply}</div>
                                                <div style={{ color: "#666" }}><strong>Max:</strong> {maxSupply}</div>
                                            </div>
                                        </td>
                                        <td style={tableCellStyle}>
                                            <code style={{ fontSize: "11px", wordBreak: "break-all" }}>
                                                {token.denom}
                                            </code>
                                        </td>
                                        <td style={tableCellStyle}>
                                            <code style={{ fontSize: "11px", wordBreak: "break-all" }}>
                                                {token.creator}
                                            </code>
                                        </td>
                                        <td style={tableCellStyle}>{founderStatus}</td>
                                        {showActions && (
                                            <td style={tableCellStyle}>{renderActions?.(token)}</td>
                                        )}
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </CardContent>
        </Card>
    );
}


