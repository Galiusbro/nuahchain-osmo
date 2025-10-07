export interface BroadcastResult {
    hash: string;
    height: number;
    gasUsed: number;
    rawLog: string;
}

export interface NetworkConfigState {
    chainId: string;
    chainName: string;
    rpcEndpoint: string;
    restEndpoint: string;
    bech32Prefix: string;
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: string;
    gasPriceLow: string;
    gasPriceAverage: string;
    gasPriceHigh: string;
}

export interface TokenFormState {
    subdenom: string;
    name: string;
    symbol: string;
    decimals: string;
    memo: string;
}

export interface TradeFormState {
    denom: string;
    paymentAmount: string;
    paymentDenom: string;
    tokenAmount: string;
    minTokens: string;
    minPrice: string;
}

export interface UserToken {
    denom: string;
    creator: string;
    name?: string;
    symbol?: string;
    description?: string;
    max_supply?: string;
    current_supply?: string;
    founder_tokens_claimed?: string;
    [key: string]: unknown;
}

export interface UserTokensResponse {
    user_tokens?: UserToken[];
    pagination?: {
        next_key?: string | null;
    };
}

export interface KeplrCurrency {
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    coinGeckoId?: string;
}

export interface KeplrCurrencyWithGas extends KeplrCurrency {
    gasPriceStep?: {
        low: number;
        average: number;
        high: number;
    };
}

export interface TradePreviewState {
    tokensUnits?: Decimal;
    payoutNUAH?: Decimal;
    loading?: boolean;
    error?: string;
}

export type TabType = "create" | "my" | "all" | "trade" | "leverage";
export type TradeMode = "buy" | "sell";


