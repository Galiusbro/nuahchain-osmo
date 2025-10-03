import { Registry } from "@cosmjs/proto-signing";
import {
    assertIsDeliverTxSuccess,
    defaultRegistryTypes,
    DeliverTxResponse,
    SigningStargateClient
} from "@cosmjs/stargate";
import Decimal from "decimal.js";
import type { ChangeEvent, CSSProperties, FormEvent, ReactNode } from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
    MSG_BUY_FOUNDER_TOKENS_TYPE_URL,
    MSG_BUY_TOKENS_TYPE_URL,
    MSG_CREATE_USER_TOKEN_TYPE_URL,
    MSG_SELL_TOKENS_TYPE_URL,
    MsgBuyFounderTokens,
    MsgBuyTokens,
    MsgCreateUserToken,
    MsgSellTokens
} from "./codec/usertoken";
import {
    MSG_OPEN_POSITION_TYPE_URL,
    MSG_CLOSE_POSITION_TYPE_URL,
    MSG_ADD_COLLATERAL_TYPE_URL,
    MsgOpenPosition,
    MsgClosePosition,
    MsgAddCollateral,
    PositionSide
} from "./codec/leverage";
import WalletBalance from "./components/WalletBalance";
import LeverageTrading from "./components/LeverageTrading";

interface BroadcastResult {
    hash: string;
    height: number;
    gasUsed: number;
    rawLog: string;
}

interface NetworkConfigState {
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

interface TokenFormState {
    subdenom: string;
    name: string;
    symbol: string;
    decimals: string;
    memo: string;
}

interface TradeFormState {
    denom: string;
    paymentAmount: string;
    paymentDenom: string;
    tokenAmount: string;
    minTokens: string;
    minPrice: string;
}

interface UserToken {
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

interface UserTokensResponse {
    user_tokens?: UserToken[];
    pagination?: {
        next_key?: string | null;
    };
}

interface KeplrCurrency {
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    coinGeckoId?: string;
}

interface KeplrCurrencyWithGas extends KeplrCurrency {
    gasPriceStep?: {
        low: number;
        average: number;
        high: number;
    };
}

interface TradePreviewState {
    tokensUnits?: Decimal;
    payoutNUAH?: Decimal;
    loading?: boolean;
    error?: string;
}

const emptyForm: TokenFormState = {
    subdenom: "",
    name: "",
    symbol: "",
    decimals: "6",
    memo: ""
};

const emptyTradeForm: TradeFormState = {
    denom: "",
    paymentAmount: "",
    paymentDenom: "unuah",
    tokenAmount: "",
    minTokens: "",
    minPrice: ""
};

const defaultNetwork: NetworkConfigState = {
    chainId: "nuahchain",
    chainName: "Nuahchain",
    rpcEndpoint: "http://localhost:26657",
    restEndpoint: "http://localhost:1317",
    bech32Prefix: "nuah",
    coinDenom: "NUAH",
    coinMinimalDenom: "unuah",
    coinDecimals: "6",
    gasPriceLow: "0.005",
    gasPriceAverage: "0.025",
    gasPriceHigh: "0.04"
};

const fieldLabelStyle: CSSProperties = {
    display: "block",
    fontWeight: 600,
    marginBottom: "4px"
};

const inputStyle: CSSProperties = {
    width: "100%",
    padding: "8px 10px",
    borderRadius: "6px",
    border: "1px solid #ccc",
    fontSize: "14px",
    marginBottom: "16px"
};

const tableHeaderCellStyle: CSSProperties = {
    fontWeight: 600,
    textAlign: "left",
    padding: "10px 12px",
    backgroundColor: "#f7fafc",
    borderBottom: "1px solid #e2e8f0"
};

const tableCellStyle: CSSProperties = {
    padding: "10px 12px",
    borderBottom: "1px solid #e2e8f0"
};

const FOUNDER_TRANCHE_AMOUNT = "10000000";
const FOUNDER_TRANCHE_AMOUNT_DISPLAY = "10,000,000";
const FOUNDER_TRANCHE_COST_DISPLAY = "500 NUAH";

const buildBech32Config = (prefix: string) => ({
    bech32PrefixAccAddr: prefix,
    bech32PrefixAccPub: `${prefix}pub`,
    bech32PrefixValAddr: `${prefix}valoper`,
    bech32PrefixValPub: `${prefix}valoperpub`,
    bech32PrefixConsAddr: `${prefix}valcons`,
    bech32PrefixConsPub: `${prefix}valconspub`
});

const formatAmount = (value?: string | null, decimals: number = 0) => {
    const raw = value && value.trim() !== "" ? value : "0";
    try {
        const big = BigInt(raw);

        if (decimals === 0) {
            // No decimals, format as-is
            return big.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
        }

        // Convert from base units to display units
        const divisor = BigInt(10 ** decimals);
        const wholePart = big / divisor;
        const fractionalPart = big % divisor;

        // Format the whole part with commas
        const wholeFormatted = wholePart.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");

        if (fractionalPart === 0n) {
            return wholeFormatted;
        }

        // Format fractional part (remove trailing zeros)
        const fractionalStr = fractionalPart.toString().padStart(decimals, '0');
        const trimmedFractional = fractionalStr.replace(/0+$/, '');

        return trimmedFractional.length > 0 ? `${wholeFormatted}.${trimmedFractional}` : wholeFormatted;
    } catch {
        return raw;
    }
};

const isTokenWithProperDecimals = (token: UserToken): boolean => {
    // Check if token has large supply values (indicating proper base units)
    // Proper tokens should have supplies in the billions+ range for 6 decimals
    try {
        const currentSupply = BigInt(token.current_supply || "0");
        const maxSupply = BigInt(token.max_supply || "0");

        // If current supply is > 1 billion, it's likely in base units
        // If it's < 1 billion, it's likely in display units (old format)
        return currentSupply > 1_000_000_000n || maxSupply > 1_000_000_000n;
    } catch {
        return false;
    }
};

Decimal.set({ precision: 40, rounding: Decimal.ROUND_FLOOR });

function App() {
    const [activeTab, setActiveTab] = useState<"create" | "my" | "all" | "trade" | "leverage">("create");
    const [network, setNetwork] = useState<NetworkConfigState>(defaultNetwork);
    const [feeDenom, setFeeDenom] = useState(defaultNetwork.coinMinimalDenom);
    const [feeAmount, setFeeAmount] = useState("25000");
    const [gasLimit, setGasLimit] = useState("1500000");
    const [form, setForm] = useState<TokenFormState>(emptyForm);
    const [tradeForm, setTradeForm] = useState<TradeFormState>(emptyTradeForm);
    const [tradeMode, setTradeMode] = useState<"buy" | "sell">("buy");
    const [client, setClient] = useState<SigningStargateClient | null>(null);
    const [walletAddress, setWalletAddress] = useState<string>("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [buyingDenom, setBuyingDenom] = useState<string | null>(null);
    const [status, setStatus] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [txResult, setTxResult] = useState<BroadcastResult | null>(null);
    const [allTokens, setAllTokens] = useState<UserToken[]>([]);
    const [isFetchingTokens, setIsFetchingTokens] = useState(false);
    const [tokensError, setTokensError] = useState<string | null>(null);
    const [myTokens, setMyTokens] = useState<UserToken[]>([]);
    const [priceCache, setPriceCache] = useState<Record<string, Decimal>>({});
    const [tradePreview, setTradePreview] = useState<TradePreviewState>({});
    const [keplrCurrenciesMap, setKeplrCurrenciesMap] = useState<Record<string, KeplrCurrency>>({});
    const [addingKeplrDenom, setAddingKeplrDenom] = useState<string | null>(null);

    const registry = useMemo(
        () =>
            new Registry([
                ...defaultRegistryTypes,
                [MSG_CREATE_USER_TOKEN_TYPE_URL, MsgCreateUserToken],
                [MSG_BUY_FOUNDER_TOKENS_TYPE_URL, MsgBuyFounderTokens],
                [MSG_BUY_TOKENS_TYPE_URL, MsgBuyTokens],
                [MSG_SELL_TOKENS_TYPE_URL, MsgSellTokens]
            ]),
        []
    );

    const restBaseUrl = useMemo(
        () => network.restEndpoint.trim().replace(/\/+$/, ""),
        [network.restEndpoint]
    );

    const baseCurrency = useMemo<KeplrCurrencyWithGas>(() => {
        const coinDenom = network.coinDenom.trim() || "NUAH";
        const coinMinimalDenom = network.coinMinimalDenom.trim() || "unuah";
        const decimalsValue = Number(network.coinDecimals);
        const coinDecimals = Number.isFinite(decimalsValue) ? decimalsValue : 6;
        const gasLow = Number(network.gasPriceLow) || 0;
        const gasAverage = Number(network.gasPriceAverage) || gasLow;
        const gasHigh = Number(network.gasPriceHigh) || gasAverage;
        return {
            coinDenom,
            coinMinimalDenom,
            coinDecimals,
            gasPriceStep: {
                low: gasLow,
                average: gasAverage,
                high: gasHigh
            }
        };
    }, [network.coinDecimals, network.coinDenom, network.coinMinimalDenom, network.gasPriceAverage, network.gasPriceHigh, network.gasPriceLow]);

    const fetchTokens = useCallback(async () => {
        if (!restBaseUrl) {
            setAllTokens([]);
            setTokensError(null);
            return;
        }

        setIsFetchingTokens(true);
        setTokensError(null);

        try {
            const tokens: UserToken[] = [];
            let nextKey: string | undefined;

            do {
                const query = nextKey ? `?pagination.key=${encodeURIComponent(nextKey)}` : "";
                const response = await fetch(`${restBaseUrl}/osmosis/usertoken/v1beta1/user_tokens${query}`);
                if (!response.ok) {
                    throw new Error(`REST query failed with status ${response.status}`);
                }

                const data: UserTokensResponse = await response.json();
                if (Array.isArray(data.user_tokens)) {
                    tokens.push(...data.user_tokens);
                }

                const rawNextKey = data.pagination?.next_key ?? undefined;
                nextKey = rawNextKey && rawNextKey !== "" ? rawNextKey : undefined;
            } while (nextKey);

            setAllTokens(tokens);
        } catch (fetchError) {
            setTokensError(fetchError instanceof Error ? fetchError.message : String(fetchError));
        } finally {
            setIsFetchingTokens(false);
        }
    }, [restBaseUrl]);

    const fetchTokensByCreator = useCallback(async (creator: string) => {
        if (!restBaseUrl || !creator) {
            return [];
        }

        try {
            const tokens: UserToken[] = [];
            let nextKey: string | undefined;

            do {
                const query = nextKey ? `?pagination.key=${encodeURIComponent(nextKey)}` : "";
                const response = await fetch(`${restBaseUrl}/osmosis/usertoken/v1beta1/user_tokens${query}`);
                if (!response.ok) {
                    throw new Error(`REST query failed with status ${response.status}`);
                }

                const data: UserTokensResponse = await response.json();

                if (Array.isArray(data.user_tokens)) {
                    const creatorTokens = data.user_tokens.filter((token) => token.creator === creator);
                    tokens.push(...creatorTokens);
                }

                const rawNextKey = data.pagination?.next_key ?? undefined;
                nextKey = rawNextKey && rawNextKey !== "" ? rawNextKey : undefined;
            } while (nextKey);

            return tokens;
        } catch (fetchError) {
            console.error('Error fetching tokens by creator:', fetchError);
            return [];
        }
    }, [restBaseUrl]);

    useEffect(() => {
        void fetchTokens();
    }, [fetchTokens]);

    const fetchPrice = useCallback(async (denom: string) => {
        if (!restBaseUrl || !denom) {
            return;
        }

        try {
            const response = await fetch(
                `${restBaseUrl}/osmosis/usertoken/v1beta1/bonding_curve_price/${encodeURI(denom)}`
            );
            if (!response.ok) {
                throw new Error(`Failed to fetch price: ${response.status}`);
            }

            const data = await response.json();
            const rawPrice = data?.price ?? data?.bonding_curve_price ?? data?.price?.price;
            if (!rawPrice) {
                throw new Error("Malformed price response");
            }

            const price = new Decimal(rawPrice);
            setPriceCache((prev) => ({ ...prev, [denom]: price }));
        } catch (priceError) {
            console.error("Price fetch error:", priceError);
            setPriceCache((prev) => ({ ...prev, [denom]: new Decimal(0) }));
        }
    }, [restBaseUrl]);

    const fetchTokenEstimate = useCallback(async (denom: string, paymentAmount: string): Promise<Decimal | null> => {
        if (!restBaseUrl || !denom || !paymentAmount) {
            return null;
        }

        try {
            // Convert payment amount to micro units (base units)
            const paymentMicro = new Decimal(paymentAmount).mul(MICRO_FACTOR);

            // Call the backend API to get accurate token estimate
            // This should use the same calculation as CalculateTokensFromPayment
            const url = `${restBaseUrl}/osmosis/usertoken/v1beta1/estimate_tokens?denom=${encodeURI(denom)}&payment_amount=${paymentMicro.toFixed(0)}`;
            const response = await fetch(url);

            if (!response.ok) {
                // If the API endpoint doesn't exist, fall back to improved calculation
                if (response.status === 404) {
                    return await calculateTokenEstimateLocally(denom, paymentAmount);
                }
                throw new Error(`Failed to fetch token estimate: ${response.status}`);
            }

            const data = await response.json();
            const tokensReceived = data?.tokens_received ?? data?.estimated_tokens;
            if (!tokensReceived) {
                throw new Error("Malformed estimate response");
            }

            // Convert from base units to display units
            return new Decimal(tokensReceived).div(MICRO_FACTOR);
        } catch (estimateError) {
            console.error("Token estimate error:", estimateError);

            // Fall back to local calculation
            return await calculateTokenEstimateLocally(denom, paymentAmount);
        }
    }, [restBaseUrl]);

    const calculateTokenEstimateLocally = useCallback(async (denom: string, paymentAmount: string): Promise<Decimal | null> => {
        try {
            // Get current token info
            const token = allTokens.find(t => t.denom === denom);
            if (!token) {
                return null;
            }

            const payment = new Decimal(paymentAmount);
            const bondingCurvePayment = payment.mul(0.3); // 30% goes to bonding curve

            // Get current price
            const price = priceCache[denom];
            if (!price || price.lte(0)) {
                return null;
            }

            // Improved calculation based on bonding curve parameters
            // Using the same logic as the backend but simplified
            const startPrice = new Decimal("0.0002"); // BondingCurveStartPrice
            const endPrice = new Decimal("1.0");      // BondingCurveEndPrice
            const maxSupply = new Decimal("30");       // 30M tokens max on curve

            // Get current supply on curve (simplified)
            const currentSupply = token.current_supply ? new Decimal(token.current_supply).div(MICRO_FACTOR) : new Decimal(0);
            const initialCirculating = new Decimal("60"); // 60M initially circulating
            const currentCurveSupply = Decimal.max(currentSupply.minus(initialCirculating), new Decimal(0));

            // For small amounts, use current price approximation
            // For larger amounts, this will still be approximate but better than before
            const tokensAtCurrentPrice = bondingCurvePayment.div(price);

            // Adjust for the fact that price increases as we buy more
            // This is a simplified approximation of the integral
            const priceIncreaseFactor = new Decimal(1.1); // Assume ~10% price increase on average
            const estimatedTokens = tokensAtCurrentPrice.div(priceIncreaseFactor);

            // Ensure we don't exceed remaining capacity
            const remainingCapacity = maxSupply.minus(currentCurveSupply);
            return Decimal.min(estimatedTokens, remainingCapacity);

        } catch (error) {
            console.error("Local token estimate error:", error);
            return null;
        }
    }, [allTokens, priceCache]);

    useEffect(() => {
        if (tradeForm.denom) {
            void fetchPrice(tradeForm.denom);
        }
    }, [tradeForm.denom, fetchPrice]);


    const handleTabChange = (tab: "create" | "my" | "all" | "trade") => {
        setActiveTab(tab);
        if (tab !== "create") {
            void fetchTokens();
        }
    };

    useEffect(() => {
        setTradeForm((current) => ({
            ...current,
            paymentDenom: network.coinMinimalDenom.trim() || "unuah"
        }));
    }, [network.coinMinimalDenom]);

    const availableBuyTokens = useMemo(() => allTokens.filter(token => isTokenWithProperDecimals(token)), [allTokens]);
    const availableSellTokens = useMemo(() => myTokens.filter(token => isTokenWithProperDecimals(token)), [myTokens]);

    useEffect(() => {
        setTradeForm({
            ...emptyTradeForm,
            paymentDenom: network.coinMinimalDenom.trim() || "unuah",
            denom:
                tradeMode === "buy"
                    ? availableBuyTokens[0]?.denom ?? ""
                    : availableSellTokens[0]?.denom ?? ""
        });
    }, [tradeMode, availableBuyTokens, availableSellTokens, network.coinMinimalDenom]);

    useEffect(() => {
        if (tradeMode === "buy" && !tradeForm.denom && availableBuyTokens.length > 0) {
            setTradeForm((current) => ({ ...current, denom: availableBuyTokens[0].denom }));
        }
        if (tradeMode === "sell" && !tradeForm.denom && availableSellTokens.length > 0) {
            setTradeForm((current) => ({ ...current, denom: availableSellTokens[0].denom }));
        }
    }, [tradeMode, tradeForm.denom, availableBuyTokens, availableSellTokens]);

    const MICRO_FACTOR = new Decimal(1_000_000);

    const parseAmountToMicro = (value: string): Decimal => {
        const trimmed = value.trim();
        if (trimmed === "") {
            return new Decimal(0);
        }
        const numeric = new Decimal(trimmed);
        if (numeric.isNegative()) {
            throw new Error("Amount cannot be negative");
        }
        return numeric.mul(MICRO_FACTOR);
    };

    const decimalToMicroString = (value: Decimal): string => {
        return value.toFixed(0, Decimal.ROUND_FLOOR);
    };

    const handleBuyTokens = async (e: FormEvent) => {
        e.preventDefault();
        if (!client || !walletAddress) {
            setError("Connect Keplr before buying tokens.");
            return;
        }

        try {
            const trimmedFeeDenom = feeDenom.trim();
            const trimmedFeeAmount = feeAmount.trim();
            const trimmedGasLimit = gasLimit.trim();

            if (!trimmedFeeDenom || !trimmedFeeAmount || !trimmedGasLimit) {
                throw new Error("Fee denom, amount, and gas limit must be configured before buying tokens.");
            }

            if (!tradeForm.denom || !tradeForm.paymentAmount) {
                throw new Error("Token denom and payment amount are required.");
            }

            const selectedToken = availableBuyTokens.find((token) => token.denom === tradeForm.denom);
            if (!selectedToken) {
                throw new Error("Selected token not found. Please choose a token from the list.");
            }

            setIsSubmitting(true);
            setStatus("Buying tokens...");
            setError(null);
            setTxResult(null);

            const paymentAmountDecimal = parseAmountToMicro(tradeForm.paymentAmount);
            const minTokensDecimal = tradeForm.minTokens ? parseAmountToMicro(tradeForm.minTokens) : new Decimal(0);

            const paymentAmountInMicroUnits = decimalToMicroString(paymentAmountDecimal);
            const minTokensInMicroUnits = decimalToMicroString(minTokensDecimal);

            const msg = {
                typeUrl: MSG_BUY_TOKENS_TYPE_URL,
                value: MsgBuyTokens.fromPartial({
                    buyer: walletAddress,
                    denom: tradeForm.denom,
                    amount: {
                        denom: tradeForm.paymentDenom,
                        amount: paymentAmountInMicroUnits
                    },
                    min_tokens: minTokensInMicroUnits
                })
            };

            const fee = {
                amount: [
                    {
                        denom: trimmedFeeDenom,
                        amount: trimmedFeeAmount
                    }
                ],
                gas: trimmedGasLimit
            };

            const response: DeliverTxResponse = await client.signAndBroadcast(walletAddress, [msg], fee);
            assertIsDeliverTxSuccess(response);

            const result: BroadcastResult = {
                hash: response.transactionHash || "",
                height: response.height,
                gasUsed: Number(response.gasUsed),
                rawLog: response.rawLog || ""
            };

            setTxResult(result);
            setStatus("Tokens purchased successfully!");
            setTradeForm(emptyTradeForm);

            // Refresh token data after purchase
            await fetchTokens();
        } catch (buyError) {
            console.error(buyError);
            setStatus(null);
            setError(buyError instanceof Error ? buyError.message : String(buyError));
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleSellTokens = async (e: FormEvent) => {
        e.preventDefault();
        if (!client || !walletAddress) {
            setError("Connect Keplr before selling tokens.");
            return;
        }

        try {
            const trimmedFeeDenom = feeDenom.trim();
            const trimmedFeeAmount = feeAmount.trim();
            const trimmedGasLimit = gasLimit.trim();

            if (!trimmedFeeDenom || !trimmedFeeAmount || !trimmedGasLimit) {
                throw new Error("Fee denom, amount, and gas limit must be configured before selling tokens.");
            }

            if (!tradeForm.denom || !tradeForm.tokenAmount) {
                throw new Error("Token denom and token amount are required.");
            }

            const selectedToken = availableSellTokens.find((token) => token.denom === tradeForm.denom);
            if (!selectedToken) {
                throw new Error("Selected token not found. Please choose a token from the list.");
            }

            setIsSubmitting(true);
            setStatus("Selling tokens...");
            setError(null);
            setTxResult(null);

            const tokenAmountDecimal = parseAmountToMicro(tradeForm.tokenAmount);
            const minPriceDecimal = tradeForm.minPrice ? parseAmountToMicro(tradeForm.minPrice) : new Decimal(0);

            const tokenAmountInMicroUnits = decimalToMicroString(tokenAmountDecimal);
            const minPriceInMicroUnits = decimalToMicroString(minPriceDecimal);

            const msg = {
                typeUrl: MSG_SELL_TOKENS_TYPE_URL,
                value: MsgSellTokens.fromPartial({
                    seller: walletAddress,
                    denom: tradeForm.denom,
                    amount: {
                        denom: tradeForm.denom,
                        amount: tokenAmountInMicroUnits
                    },
                    min_price: minPriceInMicroUnits
                })
            };

            const fee = {
                amount: [
                    {
                        denom: trimmedFeeDenom,
                        amount: trimmedFeeAmount
                    }
                ],
                gas: trimmedGasLimit
            };

            const response: DeliverTxResponse = await client.signAndBroadcast(walletAddress, [msg], fee);
            assertIsDeliverTxSuccess(response);

            const result: BroadcastResult = {
                hash: response.transactionHash || "",
                height: response.height,
                gasUsed: Number(response.gasUsed),
                rawLog: response.rawLog || ""
            };

            setTxResult(result);
            setStatus("Tokens sold successfully!");
            setTradeForm(emptyTradeForm);

            // Refresh token data after sale
            await fetchTokens();
        } catch (sellError) {
            console.error(sellError);
            setStatus(null);
            setError(sellError instanceof Error ? sellError.message : String(sellError));
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleNetworkChange = (field: keyof NetworkConfigState) => (event: ChangeEvent<HTMLInputElement>) => {
        const { value } = event.target;
        setNetwork((current) => ({ ...current, [field]: value }));

        if (field === "coinMinimalDenom") {
            setFeeDenom(value);
        }
    };

    const suggestChain = useCallback(
        async (extraCurrencies: KeplrCurrency[]) => {
            if (!window.keplr?.experimentalSuggestChain) {
                throw new Error("Keplr experimentalSuggestChain API is not available.");
            }

            const trimmedChainId = network.chainId.trim();
            if (!trimmedChainId) {
                throw new Error("Chain ID is required to update Keplr assets.");
            }

            const rpc = network.rpcEndpoint.trim();
            const rest = network.restEndpoint.trim();
            if (!rpc || !rest) {
                throw new Error("RPC and REST endpoints are required to update Keplr assets.");
            }

            // Create fee currencies with gasPriceStep for all currencies
            const allCurrencies = [baseCurrency, ...extraCurrencies];
            const feeCurrencies = allCurrencies.map(currency => ({
                ...currency,
                gasPriceStep: {
                    low: parseFloat(network.gasPriceLow) || 0.005,
                    average: parseFloat(network.gasPriceAverage) || 0.025,
                    high: parseFloat(network.gasPriceHigh) || 0.04
                }
            }));

            await window.keplr.experimentalSuggestChain({
                chainId: trimmedChainId,
                chainName: network.chainName.trim() || trimmedChainId,
                rpc,
                rest,
                bip44: { coinType: 118 },
                bech32Config: buildBech32Config(network.bech32Prefix.trim() || "nuah"),
                currencies: allCurrencies,
                feeCurrencies: feeCurrencies,
                stakeCurrency: baseCurrency,
            });
        },
        [baseCurrency, network.bech32Prefix, network.chainId, network.chainName, network.rpcEndpoint, network.restEndpoint]
    );

    const buildCurrencyFromToken = useCallback((token: UserToken): KeplrCurrency => {
        const denom = token.denom;
        const derivedSymbol = (token.symbol || token.name || denom.split("/").pop() || denom)
            .toString()
            .toUpperCase()
            .slice(0, 16);

        return {
            coinDenom: derivedSymbol,
            coinMinimalDenom: denom,
            coinDecimals: 6
        };
    }, []);

    const handleConnectWallet = async () => {
        try {
            setError(null);
            setStatus("Connecting wallet...");
            setTxResult(null);

            if (!window.keplr) {
                throw new Error("Keplr extension not detected. Please install or unlock Keplr.");
            }

            const trimmedChainId = network.chainId.trim();
            if (!trimmedChainId) {
                throw new Error("Chain ID is required.");
            }

            const rpc = network.rpcEndpoint.trim();
            if (!rpc) {
                throw new Error("RPC endpoint is required.");
            }

            const rest = network.restEndpoint.trim();
            if (!rest) {
                throw new Error("REST endpoint is required.");
            }

            const chainName = network.chainName.trim() || trimmedChainId;
            const bech32Prefix = network.bech32Prefix.trim() || "nuah";
            const coinDenom = network.coinDenom.trim() || "NUAH";
            const minimalDenom = network.coinMinimalDenom.trim();
            if (!minimalDenom) {
                throw new Error("Coin minimal denom is required.");
            }

            const decimals = Number(network.coinDecimals);
            if (!Number.isInteger(decimals) || decimals < 0 || decimals > 18) {
                throw new Error("Coin decimals must be an integer between 0 and 18.");
            }

            const gasPriceLow = Number(network.gasPriceLow);
            const gasPriceAverage = Number(network.gasPriceAverage);
            const gasPriceHigh = Number(network.gasPriceHigh);

            if ([gasPriceLow, gasPriceAverage, gasPriceHigh].some((value) => Number.isNaN(value) || value < 0)) {
                throw new Error("Gas price steps must be non-negative numbers.");
            }

            if (window.keplr.experimentalSuggestChain) {
                await window.keplr.experimentalSuggestChain({
                    chainId: trimmedChainId,
                    chainName,
                    rpc,
                    rest,
                    bip44: { coinType: 118 },
                    bech32Config: buildBech32Config(bech32Prefix),
                    stakeCurrency: {
                        coinDenom,
                        coinMinimalDenom: minimalDenom,
                        coinDecimals: decimals
                    },
                    currencies: [
                        {
                            coinDenom,
                            coinMinimalDenom: minimalDenom,
                            coinDecimals: decimals
                        }
                    ],
                    feeCurrencies: [
                        {
                            coinDenom,
                            coinMinimalDenom: minimalDenom,
                            coinDecimals: decimals,
                            gasPriceStep: {
                                low: gasPriceLow,
                                average: gasPriceAverage,
                                high: gasPriceHigh
                            }
                        }
                    ],
                    gasPriceStep: {
                        low: gasPriceLow,
                        average: gasPriceAverage,
                        high: gasPriceHigh
                    }
                });
            }

            await window.keplr.enable(trimmedChainId);
            const signer = await window.keplr.getOfflineSignerAuto(trimmedChainId);
            const accounts = await signer.getAccounts();

            if (accounts.length === 0) {
                throw new Error("No accounts returned by the wallet.");
            }

            const signingClient = await SigningStargateClient.connectWithSigner(rpc, signer, {
                registry
            });

            setClient(signingClient);
            setWalletAddress(accounts[0].address);
            setStatus(`Wallet connected: ${accounts[0].address}`);

            await fetchTokens();
            if (Object.keys(keplrCurrenciesMap).length > 0) {
                await suggestChain(Object.values(keplrCurrenciesMap));
            }
        } catch (connectError) {
            console.error(connectError);
            setClient(null);
            setWalletAddress("");
            setStatus(null);
            setTxResult(null);
            setError(connectError instanceof Error ? connectError.message : String(connectError));
        }
    };

    const handleInputChange = (field: keyof TokenFormState) => (event: ChangeEvent<HTMLInputElement>) => {
        const { value } = event.target;
        setForm((current) => ({ ...current, [field]: value }));
    };

    const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();

        if (!client || !walletAddress) {
            setError("Connect a wallet before submitting the transaction.");
            return;
        }

        try {
            setIsSubmitting(true);
            setError(null);
            setStatus("Preparing transaction...");
            setTxResult(null);

            const trimmedSubdenom = form.subdenom.trim();
            const trimmedName = form.name.trim();
            const trimmedSymbol = form.symbol.trim();

            if (!trimmedSubdenom || !trimmedName || !trimmedSymbol) {
                throw new Error("Subdenom, name, and symbol are required.");
            }

            const decimalsNumber = Number(form.decimals);
            if (!Number.isInteger(decimalsNumber) || decimalsNumber < 0 || decimalsNumber > 18) {
                throw new Error("Decimals must be an integer between 0 and 18.");
            }

            const trimmedFeeDenom = feeDenom.trim();
            const trimmedFeeAmount = feeAmount.trim();
            const trimmedGasLimit = gasLimit.trim();

            if (!trimmedFeeDenom) {
                throw new Error("Fee denom is required.");
            }

            if (!trimmedFeeAmount) {
                throw new Error("Fee amount is required.");
            }

            if (!trimmedGasLimit) {
                throw new Error("Gas limit is required.");
            }

            const msg = {
                typeUrl: MSG_CREATE_USER_TOKEN_TYPE_URL,
                value: MsgCreateUserToken.fromPartial({
                    creator: walletAddress,
                    subdenom: trimmedSubdenom,
                    name: trimmedName,
                    symbol: trimmedSymbol,
                    decimals: decimalsNumber
                })
            };

            const fee = {
                amount: [
                    {
                        denom: trimmedFeeDenom,
                        amount: trimmedFeeAmount
                    }
                ],
                gas: trimmedGasLimit
            };

            setStatus("Broadcasting transaction...");

            const response: DeliverTxResponse = await client.signAndBroadcast(
                walletAddress,
                [msg],
                fee,
                form.memo.trim() || undefined
            );

            assertIsDeliverTxSuccess(response);

            const result: BroadcastResult = {
                hash: response.transactionHash || "",
                height: response.height,
                gasUsed: Number(response.gasUsed),
                rawLog: response.rawLog || ""
            };

            setTxResult(result);
            setStatus("Token created and transaction accepted by the chain.");
            setForm((current) => ({ ...current, subdenom: "", name: "", symbol: "" }));

            // Update both all tokens and my tokens
            await fetchTokens();
            if (walletAddress) {
                const updatedMyTokens = await fetchTokensByCreator(walletAddress);
                setMyTokens(updatedMyTokens);

                // Switch to "My tokens" tab to show the newly created token
                setActiveTab("my");
            }
        } catch (submitError) {
            console.error(submitError);
            setTxResult(null);
            setStatus(null);
            setError(submitError instanceof Error ? submitError.message : String(submitError));
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleBuyFounderTokens = async (token: UserToken) => {
        if (!client || !walletAddress) {
            setError("Connect Keplr before buying founder tokens.");
            return;
        }

        try {
            const trimmedFeeDenom = feeDenom.trim();
            const trimmedFeeAmount = feeAmount.trim();
            const trimmedGasLimit = gasLimit.trim();

            if (!trimmedFeeDenom || !trimmedFeeAmount || !trimmedGasLimit) {
                throw new Error("Fee denom, amount, and gas limit must be configured before purchasing founder tokens.");
            }

            setBuyingDenom(token.denom);
            setStatus("Purchasing founder tranche...");
            setError(null);
            setTxResult(null);

            const msg = {
                typeUrl: MSG_BUY_FOUNDER_TOKENS_TYPE_URL,
                value: MsgBuyFounderTokens.fromPartial({
                    buyer: walletAddress,
                    denom: token.denom
                })
            };

            const fee = {
                amount: [
                    {
                        denom: trimmedFeeDenom,
                        amount: trimmedFeeAmount
                    }
                ],
                gas: trimmedGasLimit
            };

            const response: DeliverTxResponse = await client.signAndBroadcast(walletAddress, [msg], fee);
            assertIsDeliverTxSuccess(response);

            const result: BroadcastResult = {
                hash: response.transactionHash || "",
                height: response.height,
                gasUsed: Number(response.gasUsed),
                rawLog: response.rawLog || ""
            };

            setTxResult(result);
            setStatus("Founder tranche purchased. Tokens vested for one year.");

            // Update both all tokens and my tokens to reflect founder tranche status
            await fetchTokens();
            if (walletAddress) {
                const updatedMyTokens = await fetchTokensByCreator(walletAddress);
                setMyTokens(updatedMyTokens);
            }
        } catch (purchaseError) {
            console.error(purchaseError);
            setStatus(null);
            setError(purchaseError instanceof Error ? purchaseError.message : String(purchaseError));
        } finally {
            setBuyingDenom(null);
        }
    };

    const handleAddTokenToKeplr = async (token: UserToken) => {
        if (!window.keplr?.experimentalSuggestChain) {
            setError("Keplr does not support chain suggestions in this browser.");
            return;
        }

        try {
            setAddingKeplrDenom(token.denom);
            setError(null);

            const currency = buildCurrencyFromToken(token);
            const nextMap = { ...keplrCurrenciesMap, [token.denom]: currency };
            setKeplrCurrenciesMap(nextMap);

            await suggestChain(Object.values(nextMap));
            await window.keplr.enable(network.chainId.trim());

            setStatus(`Token ${currency.coinDenom} is now visible in Keplr.`);
        } catch (addError) {
            console.error('Error adding token to Keplr:', addError);
            setError(addError instanceof Error ? addError.message : String(addError));
        } finally {
            setAddingKeplrDenom(null);
        }
    };

    // Fetch user's tokens when wallet address changes
    useEffect(() => {
        if (walletAddress) {
            fetchTokensByCreator(walletAddress).then((tokens) => {
                setMyTokens(tokens);
            });
        } else {
            setMyTokens([]);
        }
    }, [walletAddress, fetchTokensByCreator]);

    useEffect(() => {
        const updatePreview = async () => {
            try {
                if (!tradeForm.denom) {
                    setTradePreview({});
                    return;
                }

                if (tradeMode === "buy") {
                    if (tradeForm.paymentAmount.trim() === "") {
                        setTradePreview({});
                        return;
                    }

                    const paymentNUAH = new Decimal(tradeForm.paymentAmount || "0");
                    if (paymentNUAH.lte(0)) {
                        setTradePreview({});
                        return;
                    }

                    setTradePreview({ loading: true });

                    // Use accurate backend calculation
                    const tokensUnits = await fetchTokenEstimate(tradeForm.denom, tradeForm.paymentAmount);
                    if (tokensUnits && tokensUnits.gt(0)) {
                        setTradePreview({ tokensUnits });
                    } else {
                        setTradePreview({ error: "Unable to calculate token estimate" });
                    }
                } else {
                    if (tradeForm.tokenAmount.trim() === "") {
                        setTradePreview({});
                        return;
                    }

                    const tokensUnits = new Decimal(tradeForm.tokenAmount || "0");
                    if (tokensUnits.lte(0)) {
                        setTradePreview({});
                        return;
                    }

                    // For sell operations, use current price (this could also be improved with an API call)
                    const price = priceCache[tradeForm.denom];
                    if (!price || price.lte(0)) {
                        setTradePreview({ loading: true });
                        return;
                    }

                    const payoutNUAH = tokensUnits.mul(price);
                    setTradePreview({ payoutNUAH: Decimal.max(payoutNUAH, new Decimal(0)) });
                }
            } catch (previewError) {
                setTradePreview({ error: previewError instanceof Error ? previewError.message : String(previewError) });
            }
        };

        void updatePreview();
    }, [tradeMode, tradeForm, priceCache, fetchTokenEstimate]);

    const renderTokensTable = (tokens: UserToken[], options?: { renderActions?: (token: UserToken) => ReactNode }) => {
        const showActions = Boolean(options?.renderActions);

        return (
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
                        {tokens
                            .filter(token => isTokenWithProperDecimals(token)) // Only show tokens with proper decimal handling
                            .map((token) => {
                                const founderClaimedRaw = token.founder_tokens_claimed ?? "0";
                                const founderTrancheClaimed = founderClaimedRaw !== "" && founderClaimedRaw !== "0";
                                const hasProperDecimals = isTokenWithProperDecimals(token);
                                const decimals = 6; // All user tokens use 6 decimals

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
                                            <td style={tableCellStyle}>{options?.renderActions?.(token)}</td>
                                        )}
                                    </tr>
                                );
                            })}
                    </tbody>
                </table>
            </div>
        );
    };

    const renderTokensContent = (
        tokens: UserToken[],
        emptyMessage: string,
        options?: { renderActions?: (token: UserToken) => ReactNode }
    ) => {
        if (!restBaseUrl) {
            return <div style={{ color: "#4a5568" }}>Provide a REST endpoint to load token data.</div>;
        }

        if (tokensError) {
            return <div style={{ color: "#c53030" }}>{tokensError}</div>;
        }

        if (isFetchingTokens) {
            return <div style={{ color: "#4a5568" }}>Fetching tokens...</div>;
        }

        if (tokens.length === 0) {
            return <div style={{ color: "#4a5568" }}>{emptyMessage}</div>;
        }

        return renderTokensTable(tokens, options);
    };

    const renderFounderAction = (token: UserToken) => {
        const founderClaimedRaw = token.founder_tokens_claimed ?? "0";
        const trancheAvailable = founderClaimedRaw === "" || founderClaimedRaw === "0";
        const alreadyAddedToKeplr = Boolean(keplrCurrenciesMap[token.denom]);
        const addDisabled = addingKeplrDenom === token.denom;
        const buyDisabled = !client || !walletAddress || !trancheAvailable || buyingDenom === token.denom;

        return (
            <div style={{ display: "flex", gap: "8px", flexWrap: "wrap" }}>
                <button
                    type="button"
                    onClick={() => handleAddTokenToKeplr(token)}
                    disabled={addDisabled}
                    style={{
                        padding: "8px 14px",
                        borderRadius: "8px",
                        border: "1px solid #2b6cb0",
                        backgroundColor: alreadyAddedToKeplr ? "#2b6cb0" : "transparent",
                        color: alreadyAddedToKeplr ? "white" : "#2b6cb0",
                        fontWeight: 600,
                        cursor: addDisabled ? "not-allowed" : "pointer"
                    }}
                >
                    {addingKeplrDenom === token.denom
                        ? "Adding..."
                        : alreadyAddedToKeplr
                            ? "Update in Keplr"
                            : "Add to Keplr"}
                </button>

                {trancheAvailable ? (
                    <button
                        type="button"
                        onClick={() => handleBuyFounderTokens(token)}
                        disabled={buyDisabled}
                        style={{
                            padding: "8px 14px",
                            borderRadius: "8px",
                            border: "1px solid #38a169",
                            backgroundColor: buyDisabled ? "#cbd5e0" : "#38a169",
                            color: buyDisabled ? "#4a5568" : "white",
                            fontWeight: 600,
                            cursor: buyDisabled ? "not-allowed" : "pointer"
                        }}
                    >
                        {buyingDenom === token.denom
                            ? "Processing..."
                            : `Buy ${FOUNDER_TRANCHE_AMOUNT_DISPLAY} (${FOUNDER_TRANCHE_COST_DISPLAY})`}
                    </button>
                ) : (
                    <span style={{ color: "#4a5568", alignSelf: "center" }}>Founder tranche claimed</span>
                )}
            </div>
        );
    };

    const handleRefreshTokens = () => {
        void fetchTokens();
    };

    return (
        <main
            style={{
                fontFamily: "Inter, system-ui, sans-serif",
                margin: "0 auto",
                maxWidth: "920px",
                padding: "32px"
            }}
        >
            <header style={{ marginBottom: "24px" }}>
                <h1 style={{ marginBottom: "8px" }}>User Token Console</h1>
                <p style={{ color: "#4a5568", lineHeight: 1.5 }}>
                    Configure your network, connect Keplr, create user tokens, and review existing denoms.
                </p>
                <nav style={{ display: "flex", gap: "12px", marginTop: "24px", flexWrap: "wrap" }}>
                    <button type="button" onClick={() => handleTabChange("create")} style={{
                        padding: "10px 18px",
                        borderRadius: "20px",
                        border: "1px solid #2b6cb0",
                        backgroundColor: activeTab === "create" ? "#2b6cb0" : "transparent",
                        color: activeTab === "create" ? "white" : "#2b6cb0",
                        fontWeight: 600,
                        cursor: "pointer"
                    }}>
                        Create token
                    </button>
                    <button type="button" onClick={() => handleTabChange("my")} style={{
                        padding: "10px 18px",
                        borderRadius: "20px",
                        border: "1px solid #2b6cb0",
                        backgroundColor: activeTab === "my" ? "#2b6cb0" : "transparent",
                        color: activeTab === "my" ? "white" : "#2b6cb0",
                        fontWeight: 600,
                        cursor: "pointer"
                    }}>
                        My tokens
                    </button>
                    <button type="button" onClick={() => handleTabChange("all")} style={{
                        padding: "10px 18px",
                        borderRadius: "20px",
                        border: "1px solid #2b6cb0",
                        backgroundColor: activeTab === "all" ? "#2b6cb0" : "transparent",
                        color: activeTab === "all" ? "white" : "#2b6cb0",
                        fontWeight: 600,
                        cursor: "pointer"
                    }}>
                        All tokens
                    </button>
                    <button type="button" onClick={() => handleTabChange("trade")} style={{
                        padding: "10px 18px",
                        borderRadius: "20px",
                        border: "1px solid #2b6cb0",
                        backgroundColor: activeTab === "trade" ? "#2b6cb0" : "transparent",
                        color: activeTab === "trade" ? "white" : "#2b6cb0",
                        fontWeight: 600,
                        cursor: "pointer"
                    }}>
                        Trade tokens
                    </button>
                    <button type="button" onClick={() => handleTabChange("leverage")} style={{
                        padding: "10px 18px",
                        borderRadius: "20px",
                        border: "1px solid #ff9800",
                        backgroundColor: activeTab === "leverage" ? "#ff9800" : "transparent",
                        color: activeTab === "leverage" ? "white" : "#ff9800",
                        fontWeight: 600,
                        cursor: "pointer"
                    }}>
                        🚀 Leverage (100x)
                    </button>
                </nav>
            </header>

            <section style={{ marginBottom: "32px", padding: "24px", border: "1px solid #e2e8f0", borderRadius: "12px" }}>
                <h2 style={{ marginBottom: "16px", fontSize: "18px" }}>Network configuration</h2>

                <div style={{ display: "flex", gap: "16px", flexWrap: "wrap" }}>
                    <div style={{ flex: "1 1 260px" }}>
                        <label style={fieldLabelStyle} htmlFor="chain-id">Chain ID</label>
                        <input
                            id="chain-id"
                            style={inputStyle}
                            value={network.chainId}
                            onChange={handleNetworkChange("chainId")}
                            placeholder="nuahchain"
                            required
                        />
                    </div>
                    <div style={{ flex: "1 1 260px" }}>
                        <label style={fieldLabelStyle} htmlFor="chain-name">Chain name</label>
                        <input
                            id="chain-name"
                            style={inputStyle}
                            value={network.chainName}
                            onChange={handleNetworkChange("chainName")}
                            placeholder="Nuahchain"
                        />
                    </div>
                    <div style={{ flex: "1 1 260px" }}>
                        <label style={fieldLabelStyle} htmlFor="bech32-prefix">Bech32 prefix</label>
                        <input
                            id="bech32-prefix"
                            style={inputStyle}
                            value={network.bech32Prefix}
                            onChange={handleNetworkChange("bech32Prefix")}
                            placeholder="nuah"
                            required
                        />
                    </div>
                </div>

                <div style={{ display: "flex", gap: "16px", flexWrap: "wrap" }}>
                    <div style={{ flex: "1 1 320px" }}>
                        <label style={fieldLabelStyle} htmlFor="rpc-endpoint">RPC endpoint</label>
                        <input
                            id="rpc-endpoint"
                            style={inputStyle}
                            value={network.rpcEndpoint}
                            onChange={handleNetworkChange("rpcEndpoint")}
                            placeholder="https://rpc.nuahchain.example.com"
                            required
                        />
                    </div>
                    <div style={{ flex: "1 1 320px" }}>
                        <label style={fieldLabelStyle} htmlFor="rest-endpoint">REST endpoint</label>
                        <input
                            id="rest-endpoint"
                            style={inputStyle}
                            value={network.restEndpoint}
                            onChange={handleNetworkChange("restEndpoint")}
                            placeholder="https://lcd.nuahchain.example.com"
                            required
                        />
                    </div>
                </div>

                <div style={{ display: "flex", gap: "16px", flexWrap: "wrap" }}>
                    <div style={{ flex: "1 1 220px" }}>
                        <label style={fieldLabelStyle} htmlFor="coin-denom">Display denom</label>
                        <input
                            id="coin-denom"
                            style={inputStyle}
                            value={network.coinDenom}
                            onChange={handleNetworkChange("coinDenom")}
                            placeholder="NUAH"
                            required
                        />
                    </div>
                    <div style={{ flex: "1 1 220px" }}>
                        <label style={fieldLabelStyle} htmlFor="coin-minimal-denom">Minimal denom</label>
                        <input
                            id="coin-minimal-denom"
                            style={inputStyle}
                            value={network.coinMinimalDenom}
                            onChange={handleNetworkChange("coinMinimalDenom")}
                            placeholder="unuah"
                            required
                        />
                    </div>
                    <div style={{ flex: "1 1 160px" }}>
                        <label style={fieldLabelStyle} htmlFor="coin-decimals">Decimals</label>
                        <input
                            id="coin-decimals"
                            style={inputStyle}
                            type="number"
                            min={0}
                            max={18}
                            value={network.coinDecimals}
                            onChange={handleNetworkChange("coinDecimals")}
                            required
                        />
                    </div>
                </div>

                <div style={{ display: "flex", gap: "16px", flexWrap: "wrap" }}>
                    <div style={{ flex: "1 1 160px" }}>
                        <label style={fieldLabelStyle} htmlFor="gas-price-low">Gas price (low)</label>
                        <input
                            id="gas-price-low"
                            style={inputStyle}
                            value={network.gasPriceLow}
                            onChange={handleNetworkChange("gasPriceLow")}
                            placeholder="0.005"
                            required
                        />
                    </div>
                    <div style={{ flex: "1 1 160px" }}>
                        <label style={fieldLabelStyle} htmlFor="gas-price-average">Gas price (average)</label>
                        <input
                            id="gas-price-average"
                            style={inputStyle}
                            value={network.gasPriceAverage}
                            onChange={handleNetworkChange("gasPriceAverage")}
                            placeholder="0.025"
                            required
                        />
                    </div>
                    <div style={{ flex: "1 1 160px" }}>
                        <label style={fieldLabelStyle} htmlFor="gas-price-high">Gas price (high)</label>
                        <input
                            id="gas-price-high"
                            style={inputStyle}
                            value={network.gasPriceHigh}
                            onChange={handleNetworkChange("gasPriceHigh")}
                            placeholder="0.04"
                            required
                        />
                    </div>
                </div>

                <div style={{ display: "flex", gap: "16px", flexWrap: "wrap" }}>
                    <div style={{ flex: "1 1 220px" }}>
                        <label style={fieldLabelStyle} htmlFor="fee-denom">Fee denom</label>
                        <input
                            id="fee-denom"
                            style={inputStyle}
                            value={feeDenom}
                            onChange={(event) => setFeeDenom(event.target.value)}
                            placeholder="unuah"
                            required
                        />
                    </div>
                    <div style={{ flex: "1 1 220px" }}>
                        <label style={fieldLabelStyle} htmlFor="fee-amount">Fee amount</label>
                        <input
                            id="fee-amount"
                            style={inputStyle}
                            value={feeAmount}
                            onChange={(event) => setFeeAmount(event.target.value)}
                            placeholder="25000"
                            required
                        />
                    </div>
                    <div style={{ flex: "1 1 220px" }}>
                        <label style={fieldLabelStyle} htmlFor="gas-limit">Gas limit</label>
                        <input
                            id="gas-limit"
                            style={inputStyle}
                            value={gasLimit}
                            onChange={(event) => setGasLimit(event.target.value)}
                            placeholder="1500000"
                            required
                        />
                    </div>
                </div>

                <button
                    type="button"
                    onClick={handleConnectWallet}
                    style={{
                        padding: "10px 18px",
                        borderRadius: "8px",
                        border: "none",
                        backgroundColor: "#2b6cb0",
                        color: "white",
                        fontWeight: 600,
                        cursor: "pointer"
                    }}
                >
                    {walletAddress ? "Reconnect" : "Connect Keplr"}
                </button>
            </section>

            {/* Wallet Balance Component */}
            <WalletBalance
                client={client}
                walletAddress={walletAddress}
                restEndpoint={network.restEndpoint}
                onError={(error) => setError(error)}
            />

            {activeTab === "create" && (
                <section style={{ padding: "24px", border: "1px solid #e2e8f0", borderRadius: "12px" }}>
                    <h2 style={{ marginBottom: "16px", fontSize: "18px" }}>Token metadata</h2>
                    <form onSubmit={handleSubmit}>
                        <label style={fieldLabelStyle} htmlFor="subdenom">
                            Subdenom
                        </label>
                        <input
                            id="subdenom"
                            style={inputStyle}
                            value={form.subdenom}
                            onChange={handleInputChange("subdenom")}
                            placeholder="mytoken"
                            required
                        />

                        <label style={fieldLabelStyle} htmlFor="name">
                            Name
                        </label>
                        <input
                            id="name"
                            style={inputStyle}
                            value={form.name}
                            onChange={handleInputChange("name")}
                            placeholder="My Token"
                            required
                        />

                        <label style={fieldLabelStyle} htmlFor="symbol">
                            Symbol
                        </label>
                        <input
                            id="symbol"
                            style={inputStyle}
                            value={form.symbol}
                            onChange={handleInputChange("symbol")}
                            placeholder="MYT"
                            required
                        />

                        <label style={fieldLabelStyle} htmlFor="decimals">
                            Decimals (0-18)
                        </label>
                        <input
                            id="decimals"
                            style={inputStyle}
                            type="number"
                            min={0}
                            max={18}
                            value={form.decimals}
                            onChange={handleInputChange("decimals")}
                            required
                        />

                        <label style={fieldLabelStyle} htmlFor="memo">
                            Memo (optional)
                        </label>
                        <input
                            id="memo"
                            style={inputStyle}
                            value={form.memo}
                            onChange={handleInputChange("memo")}
                            placeholder="Optional memo"
                        />

                        <button
                            type="submit"
                            disabled={isSubmitting || !walletAddress}
                            style={{
                                padding: "12px 18px",
                                borderRadius: "8px",
                                border: "none",
                                backgroundColor: isSubmitting || !walletAddress ? "#a0aec0" : "#38a169",
                                color: "white",
                                fontWeight: 600,
                                cursor: isSubmitting || !walletAddress ? "not-allowed" : "pointer",
                                width: "100%"
                            }}
                        >
                            {isSubmitting ? "Broadcasting..." : "Create user token"}
                        </button>
                    </form>
                </section>
            )}

            {activeTab === "my" && (
                <section style={{ padding: "24px", border: "1px solid #e2e8f0", borderRadius: "12px" }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "16px" }}>
                        <h2 style={{ fontSize: "18px", margin: 0 }}>My tokens</h2>
                        <button
                            type="button"
                            onClick={handleRefreshTokens}
                            style={{
                                padding: "8px 14px",
                                borderRadius: "8px",
                                border: "1px solid #2b6cb0",
                                backgroundColor: "transparent",
                                color: "#2b6cb0",
                                fontWeight: 600,
                                cursor: "pointer"
                            }}
                        >
                            Refresh
                        </button>
                    </div>
                    {!walletAddress ? (
                        <div style={{ color: "#4a5568" }}>Connect Keplr to view tokens you created.</div>
                    ) : (
                        renderTokensContent(myTokens, "No tokens created with this address yet.", {
                            renderActions: renderFounderAction
                        })
                    )}
                </section>
            )}

            {activeTab === "all" && (
                <section style={{ padding: "24px", border: "1px solid #e2e8f0", borderRadius: "12px" }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "16px" }}>
                        <h2 style={{ fontSize: "18px", margin: 0 }}>All user tokens</h2>
                        <button
                            type="button"
                            onClick={handleRefreshTokens}
                            style={{
                                padding: "8px 14px",
                                borderRadius: "8px",
                                border: "1px solid #2b6cb0",
                                backgroundColor: "transparent",
                                color: "#2b6cb0",
                                fontWeight: 600,
                                cursor: "pointer"
                            }}
                        >
                            Refresh
                        </button>
                    </div>
                    {renderTokensContent(allTokens, "No user tokens have been created yet.")}
                </section>
            )}

            {activeTab === "trade" && (
                <section style={{ padding: "24px", border: "1px solid #e2e8f0", borderRadius: "12px" }}>
                    <h2 style={{ fontSize: "18px", marginBottom: "16px" }}>Trade tokens</h2>

                    <div style={{ marginBottom: "24px" }}>
                        <div style={{ display: "flex", gap: "12px", marginBottom: "16px" }}>
                            <button
                                type="button"
                                onClick={() => setTradeMode("buy")}
                                style={{
                                    padding: "8px 16px",
                                    borderRadius: "8px",
                                    border: "1px solid #2b6cb0",
                                    backgroundColor: tradeMode === "buy" ? "#2b6cb0" : "transparent",
                                    color: tradeMode === "buy" ? "white" : "#2b6cb0",
                                    fontWeight: 600,
                                    cursor: "pointer"
                                }}
                            >
                                Buy Tokens
                            </button>
                            <button
                                type="button"
                                onClick={() => setTradeMode("sell")}
                                style={{
                                    padding: "8px 16px",
                                    borderRadius: "8px",
                                    border: "1px solid #2b6cb0",
                                    backgroundColor: tradeMode === "sell" ? "#2b6cb0" : "transparent",
                                    color: tradeMode === "sell" ? "white" : "#2b6cb0",
                                    fontWeight: 600,
                                    cursor: "pointer"
                                }}
                            >
                                Sell Tokens
                            </button>
                        </div>

                        {(tradeMode === "buy" ? availableBuyTokens : availableSellTokens).length === 0 ? (
                            <div style={{ color: "#4a5568", marginBottom: "16px" }}>
                                {tradeMode === "buy"
                                    ? "No user tokens available yet. Create or wait for a token to trade."
                                    : "You do not hold any user tokens to sell."}
                            </div>
                        ) : null}

                        <form onSubmit={tradeMode === "buy" ? handleBuyTokens : handleSellTokens}>
                            <div>
                                <label style={fieldLabelStyle}>Token Denom</label>
                                <select
                                    value={tradeForm.denom}
                                    onChange={(e) => setTradeForm({ ...tradeForm, denom: e.target.value })}
                                    style={{ ...inputStyle, padding: "8px 10px" }}
                                    required
                                >
                                    <option value="" disabled>
                                        {tradeMode === "buy" ? "Select a token to buy" : "Select a token to sell"}
                                    </option>
                                    {(tradeMode === "buy" ? availableBuyTokens : availableSellTokens).map((token) => (
                                        <option key={token.denom} value={token.denom}>
                                            {(token.symbol || token.name || token.denom) as string}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            {tradeMode === "buy" ? (
                                <>
                                    <div>
                                        <label style={fieldLabelStyle}>Payment Amount (NUAH)</label>
                                        <input
                                            type="text"
                                            value={tradeForm.paymentAmount}
                                            onChange={(e) => setTradeForm({ ...tradeForm, paymentAmount: e.target.value })}
                                            placeholder="Amount to spend"
                                            style={inputStyle}
                                            required
                                        />
                                    </div>
                                    <div>
                                        <label style={fieldLabelStyle}>Payment Denom</label>
                                        <input
                                            type="text"
                                            value={baseCurrency.coinDenom}
                                            readOnly
                                            style={{ ...inputStyle, backgroundColor: "#edf2f7" }}
                                        />
                                    </div>
                                    <div>
                                        <label style={fieldLabelStyle}>Minimum Tokens to Receive</label>
                                        <input
                                            type="text"
                                            value={tradeForm.minTokens}
                                            onChange={(e) => setTradeForm({ ...tradeForm, minTokens: e.target.value })}
                                            placeholder="Minimum tokens (optional)"
                                            style={inputStyle}
                                        />
                                    </div>
                                </>
                            ) : (
                                <>
                                    <div>
                                        <label style={fieldLabelStyle}>Token Amount to Sell</label>
                                        <input
                                            type="text"
                                            value={tradeForm.tokenAmount}
                                            onChange={(e) => setTradeForm({ ...tradeForm, tokenAmount: e.target.value })}
                                            placeholder="Amount of tokens to sell"
                                            style={inputStyle}
                                            required
                                        />
                                    </div>
                                    <div>
                                        <label style={fieldLabelStyle}>Minimum Price to Receive (NUAH)</label>
                                        <input
                                            type="text"
                                            value={tradeForm.minPrice}
                                            onChange={(e) => setTradeForm({ ...tradeForm, minPrice: e.target.value })}
                                            placeholder="Minimum price (optional)"
                                            style={inputStyle}
                                        />
                                    </div>
                                </>
                            )}

                            <div
                                style={{
                                    marginBottom: "16px",
                                    padding: "12px",
                                    borderRadius: "8px",
                                    border: "1px solid #e2e8f0",
                                    backgroundColor: "#f7fafc",
                                    color: tradePreview.error ? "#c53030" : "#2d3748"
                                }}
                            >
                                {tradePreview.error ? (
                                    <span>{tradePreview.error}</span>
                                ) : tradePreview.loading ? (
                                    <span>Loading preview...</span>
                                ) : tradeMode === "buy" ? (
                                    <div>
                                        <div style={{ fontWeight: 600, marginBottom: "4px" }}>Estimated tokens (approximate)</div>
                                        <div style={{ marginBottom: "6px" }}>
                                            {tradePreview.tokensUnits
                                                ? `${tradePreview.tokensUnits.toFixed(6)} (${formatAmount(
                                                    tradePreview.tokensUnits.mul(MICRO_FACTOR).toFixed(0)
                                                )} micro-units)`
                                                : "Enter amount to preview"}
                                        </div>
                                        {tradePreview.tokensUnits && (
                                            <div style={{ fontSize: "12px", color: "#666", fontStyle: "italic" }}>
                                                ⚠️ Actual amount may vary due to bonding curve dynamics
                                            </div>
                                        )}
                                    </div>
                                ) : (
                                    <div>
                                        <div style={{ fontWeight: 600, marginBottom: "4px" }}>Estimated payout</div>
                                        <div>
                                            {tradePreview.payoutNUAH
                                                ? `${tradePreview.payoutNUAH.toFixed(6)} NUAH (${formatAmount(
                                                    tradePreview.payoutNUAH.mul(MICRO_FACTOR).toFixed(0)
                                                )} unuah)`
                                                : "Enter amount to preview"}
                                        </div>
                                    </div>
                                )}
                            </div>

                            <button
                                type="submit"
                                disabled={isSubmitting || !walletAddress}
                                style={{
                                    width: "100%",
                                    padding: "12px",
                                    borderRadius: "8px",
                                    border: "none",
                                    backgroundColor: isSubmitting || !walletAddress ? "#a0aec0" : "#2b6cb0",
                                    color: "white",
                                    fontWeight: 600,
                                    cursor: isSubmitting || !walletAddress ? "not-allowed" : "pointer"
                                }}
                            >
                                {isSubmitting ? "Processing..." : tradeMode === "buy" ? "Buy Tokens" : "Sell Tokens"}
                            </button>
                        </form>
                    </div>
                </section>
            )}

            <section style={{ marginTop: "24px" }}>
                {status && (
                    <div style={{ marginBottom: "12px", color: "#2f855a", fontWeight: 600 }}>{status}</div>
                )}
                {error && (
                    <div style={{ marginBottom: "12px", color: "#c53030", whiteSpace: "pre-wrap" }}>{error}</div>
                )}
                {txResult && (
                    <div
                        style={{
                            backgroundColor: "#f0fff4",
                            border: "1px solid #9ae6b4",
                            borderRadius: "10px",
                            padding: "16px",
                            fontSize: "14px"
                        }}
                    >
                        <div><strong>Tx hash:</strong> {txResult.hash}</div>
                        <div><strong>Height:</strong> {txResult.height}</div>
                        <div><strong>Gas used:</strong> {txResult.gasUsed}</div>
                        <div style={{ marginTop: "8px" }}>
                            <strong>Raw log:</strong>
                            <pre style={{ whiteSpace: "pre-wrap", overflowX: "auto", marginTop: "6px" }}>{txResult.rawLog}</pre>
                        </div>
                    </div>
                )}
            </section>
            )}

            {activeTab === "leverage" && (
                <LeverageTrading
                    walletAddress={walletAddress || ""}
                    restEndpoint={restBaseUrl || ""}
                    onOpenPosition={async (params) => {
                        if (!client || !walletAddress) {
                            throw new Error("Wallet not connected");
                        }

                        const registry = new Registry([
                            ...defaultRegistryTypes,
                            [MSG_OPEN_POSITION_TYPE_URL, MsgOpenPosition]
                        ]);

                        const clientWithRegistry = await SigningStargateClient.connectWithSigner(
                            network.rpcEndpoint,
                            offlineSigner!,
                            { registry }
                        );

                        const msg: MsgOpenPosition = {
                            trader: walletAddress,
                            tokenDenom: params.tokenDenom,
                            collateral: {
                                denom: params.collateralDenom,
                                amount: params.collateralAmount
                            },
                            leverage: params.leverage,
                            side: params.side,
                            minPrice: params.minPrice,
                            maxPrice: params.maxPrice
                        };

                        const fee = {
                            amount: [{ denom: feeDenom, amount: feeAmount }],
                            gas: gasLimit
                        };

                        const result = await clientWithRegistry.signAndBroadcast(
                            walletAddress,
                            [{
                                typeUrl: MSG_OPEN_POSITION_TYPE_URL,
                                value: msg
                            }],
                            fee,
                            "Open leverage position"
                        );

                        assertIsDeliverTxSuccess(result);
                        setTxResult({
                            hash: result.transactionHash,
                            height: result.height,
                            gasUsed: result.gasUsed,
                            rawLog: result.rawLog
                        });
                    }}
                    onClosePosition={async (positionId, minPrice, maxPrice) => {
                        if (!client || !walletAddress) {
                            throw new Error("Wallet not connected");
                        }

                        const registry = new Registry([
                            ...defaultRegistryTypes,
                            [MSG_CLOSE_POSITION_TYPE_URL, MsgClosePosition]
                        ]);

                        const clientWithRegistry = await SigningStargateClient.connectWithSigner(
                            network.rpcEndpoint,
                            offlineSigner!,
                            { registry }
                        );

                        const msg: MsgClosePosition = {
                            trader: walletAddress,
                            positionId,
                            minPrice,
                            maxPrice
                        };

                        const fee = {
                            amount: [{ denom: feeDenom, amount: feeAmount }],
                            gas: gasLimit
                        };

                        const result = await clientWithRegistry.signAndBroadcast(
                            walletAddress,
                            [{
                                typeUrl: MSG_CLOSE_POSITION_TYPE_URL,
                                value: msg
                            }],
                            fee,
                            "Close leverage position"
                        );

                        assertIsDeliverTxSuccess(result);
                        setTxResult({
                            hash: result.transactionHash,
                            height: result.height,
                            gasUsed: result.gasUsed,
                            rawLog: result.rawLog
                        });
                    }}
                    onAddCollateral={async (positionId, amount, denom) => {
                        if (!client || !walletAddress) {
                            throw new Error("Wallet not connected");
                        }

                        const registry = new Registry([
                            ...defaultRegistryTypes,
                            [MSG_ADD_COLLATERAL_TYPE_URL, MsgAddCollateral]
                        ]);

                        const clientWithRegistry = await SigningStargateClient.connectWithSigner(
                            network.rpcEndpoint,
                            offlineSigner!,
                            { registry }
                        );

                        const msg: MsgAddCollateral = {
                            trader: walletAddress,
                            positionId,
                            amount: {
                                denom,
                                amount
                            }
                        };

                        const fee = {
                            amount: [{ denom: feeDenom, amount: feeAmount }],
                            gas: gasLimit
                        };

                        const result = await clientWithRegistry.signAndBroadcast(
                            walletAddress,
                            [{
                                typeUrl: MSG_ADD_COLLATERAL_TYPE_URL,
                                value: msg
                            }],
                            fee,
                            "Add collateral to position"
                        );

                        assertIsDeliverTxSuccess(result);
                        setTxResult({
                            hash: result.transactionHash,
                            height: result.height,
                            gasUsed: result.gasUsed,
                            rawLog: result.rawLog
                        });
                    }}
                />
            )}
        </main>
    );
}

export default App;
