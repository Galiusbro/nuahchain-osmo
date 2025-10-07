import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import type {
    BroadcastResult,
    NetworkConfigState,
    TabType,
    TokenFormState,
    TradeFormState,
    UserToken
} from '@/types';
import { Registry } from "@cosmjs/proto-signing";
import {
    assertIsDeliverTxSuccess,
    defaultRegistryTypes,
    DeliverTxResponse
} from "@cosmjs/stargate";
import Decimal from "decimal.js";
import { Coins, Eye, Plus, TrendingUp, Wallet } from 'lucide-react';
import type { ChangeEvent, CSSProperties, FormEvent, ReactNode } from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
    MSG_ADD_COLLATERAL_TYPE_URL,
    MSG_CLOSE_POSITION_TYPE_URL,
    MSG_LIQUIDATE_POSITION_TYPE_URL,
    MSG_OPEN_POSITION_TYPE_URL,
    MSG_PROVIDE_LIQUIDITY_TYPE_URL,
    MSG_REMOVE_COLLATERAL_TYPE_URL
} from "./codec/leverage";
import {
    MsgAddCollateral,
    MsgClosePosition,
    MsgLiquidatePosition,
    MsgOpenPosition,
    MsgProvideLiquidity,
    MsgRemoveCollateral
} from "./codec/leverage_proto";
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
import CreateTokenForm from './components/CreateTokenForm';
import LeverageForm from './components/LeverageForm';
import NetworkConfiguration from './components/NetworkConfiguration';
import TokenActions from './components/TokenActions';
import TokenTable from './components/TokenTable';
import TradeForm from './components/TradeForm';
import WalletBalance from './components/WalletBalance';
import { useNetwork } from './hooks/useNetwork';
import { useTokens } from './hooks/useTokens';
import { useTrading } from './hooks/useTrading';
import { useWallet } from './hooks/useWallet';

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
    rpcEndpoint: "http://localhost:5173/api",
    restEndpoint: "http://localhost:5173/rest",
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
    const [activeTab, setActiveTab] = useState<TabType>("create");
    const { network, handleNetworkChange } = useNetwork();
    const { client, walletAddress, isConnecting, connectWallet } = useWallet(network);

    // Create signAndBroadcast function for leverage module
    const signAndBroadcast = useCallback(async (messages: any[], fee: any, memo?: string) => {
        if (!client || !walletAddress) {
            throw new Error("Wallet not connected");
        }
        return await client.signAndBroadcast(walletAddress, messages, fee, memo);
    }, [client, walletAddress]);

    const [feeDenom, setFeeDenom] = useState(network.coinMinimalDenom);
    const [feeAmount, setFeeAmount] = useState("25000");
    const [gasLimit, setGasLimit] = useState("1500000");
    const [form, setForm] = useState<TokenFormState>(emptyForm);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [buyingDenom, setBuyingDenom] = useState<string | null>(null);
    const [status, setStatus] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [txResult, setTxResult] = useState<BroadcastResult | null>(null);
    const [keplrCurrenciesMap, setKeplrCurrenciesMap] = useState<Record<string, KeplrCurrency>>({});
    const [addingKeplrDenom, setAddingKeplrDenom] = useState<string | null>(null);

    const restBaseUrl = useMemo(
        () => network.restEndpoint.trim().replace(/\/+$/, ""),
        [network.restEndpoint]
    );

    const {
        allTokens,
        myTokens,
        isFetchingTokens,
        tokensError,
        priceCache,
        fetchTokens,
        fetchTokensByCreator,
        fetchPrice,
        setAllTokens,
        setMyTokens
    } = useTokens({ restBaseUrl, walletAddress });

    const {
        tradeForm,
        tradeMode,
        tradePreview,
        availableBuyTokens,
        availableSellTokens,
        handleTradeModeChange,
        handleTradeInputChange,
        setTradeForm
    } = useTrading({
        allTokens,
        myTokens,
        priceCache,
        fetchPrice,
        fetchTokenEstimate: () => Promise.resolve(null) // TODO: Implement
    });

    const registry = useMemo(
        () =>
            new Registry([
                ...defaultRegistryTypes,
                [MSG_CREATE_USER_TOKEN_TYPE_URL, MsgCreateUserToken],
                [MSG_BUY_FOUNDER_TOKENS_TYPE_URL, MsgBuyFounderTokens],
                [MSG_BUY_TOKENS_TYPE_URL, MsgBuyTokens],
                [MSG_SELL_TOKENS_TYPE_URL, MsgSellTokens],
                [MSG_OPEN_POSITION_TYPE_URL, MsgOpenPosition],
                [MSG_CLOSE_POSITION_TYPE_URL, MsgClosePosition],
                [MSG_ADD_COLLATERAL_TYPE_URL, MsgAddCollateral],
                [MSG_REMOVE_COLLATERAL_TYPE_URL, MsgRemoveCollateral],
                [MSG_LIQUIDATE_POSITION_TYPE_URL, MsgLiquidatePosition],
                [MSG_PROVIDE_LIQUIDITY_TYPE_URL, MsgProvideLiquidity]
            ]),
        []
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



    const handleTabChange = (tab: TabType) => {
        setActiveTab(tab);
        if (tab !== "create") {
            void fetchTokens();
        }
    };

    const handleConnectWallet = async () => {
        try {
            setError(null);
            setStatus("Connecting wallet...");
            setTxResult(null);

            const result = await connectWallet();
            setStatus(`Wallet connected: ${result.address}`);

            await fetchTokens();
            if (Object.keys(keplrCurrenciesMap).length > 0) {
                // await suggestChain(Object.values(keplrCurrenciesMap));
            }
        } catch (connectError) {
            console.error(connectError);
            setStatus(null);
            setTxResult(null);
            setError(connectError instanceof Error ? connectError.message : String(connectError));
        }
    };

    useEffect(() => {
        setTradeForm((current) => ({
            ...current,
            paymentDenom: network.coinMinimalDenom.trim() || "unuah"
        }));
    }, [network.coinMinimalDenom]);


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

    const handleNetworkChangeWithFee = (field: keyof NetworkConfigState) => (event: ChangeEvent<HTMLInputElement>) => {
        const { value } = event.target;
        handleNetworkChange(field)(event);

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
        <main className="min-h-screen bg-background">
            <div className="container mx-auto max-w-6xl p-6">
                <header className="mb-8">
                    <h1 className="text-4xl font-bold mb-2">User Token Console</h1>
                    <p className="text-muted-foreground text-lg">
                        Configure your network, connect Keplr, create user tokens, and review existing denoms.
                    </p>

                    <Tabs value={activeTab} onValueChange={(value) => handleTabChange(value as "create" | "my" | "all" | "trade" | "leverage")} className="mt-6">
                        <TabsList className="grid w-full grid-cols-5">
                            <TabsTrigger value="create" className="flex items-center gap-2">
                                <Plus className="h-4 w-4" />
                                Create token
                            </TabsTrigger>
                            <TabsTrigger value="my" className="flex items-center gap-2">
                                <Wallet className="h-4 w-4" />
                                My tokens
                            </TabsTrigger>
                            <TabsTrigger value="all" className="flex items-center gap-2">
                                <Eye className="h-4 w-4" />
                                All tokens
                            </TabsTrigger>
                            <TabsTrigger value="trade" className="flex items-center gap-2">
                                <Coins className="h-4 w-4" />
                                Trade tokens
                            </TabsTrigger>
                            <TabsTrigger value="leverage" className="flex items-center gap-2">
                                <TrendingUp className="h-4 w-4" />
                                🚀 Leverage (100x)
                            </TabsTrigger>
                        </TabsList>

                        <TabsContent value="create">
                            <NetworkConfiguration
                                network={network}
                                onNetworkChange={handleNetworkChangeWithFee}
                                onConnectWallet={handleConnectWallet}
                                isConnecting={isConnecting}
                                walletAddress={walletAddress}
                                status={status}
                                error={error}
                            />

                            {/* Wallet Balance */}
                            {walletAddress && (
                                <WalletBalance
                                    client={client}
                                    walletAddress={walletAddress}
                                    restEndpoint={network.restEndpoint}
                                    onError={setError}
                                />
                            )}

                            <CreateTokenForm
                                form={form}
                                onInputChange={handleInputChange}
                                onSubmit={handleSubmit}
                                isSubmitting={isSubmitting}
                                walletAddress={walletAddress}
                            />
                        </TabsContent>

                        <TabsContent value="my">
                            <TokenTable
                                tokens={myTokens}
                                emptyMessage="You haven't created any tokens yet."
                                isLoading={isFetchingTokens}
                                error={tokensError}
                                restBaseUrl={restBaseUrl}
                                renderActions={(token) => (
                                    <TokenActions
                                        token={token}
                                        onAddToKeplr={handleAddTokenToKeplr}
                                        onBuyFounderTokens={handleBuyFounderTokens}
                                        isAddingToKeplr={addingKeplrDenom === token.denom}
                                        isBuyingFounder={buyingDenom === token.denom}
                                        isAlreadyAddedToKeplr={Boolean(keplrCurrenciesMap[token.denom])}
                                        isFounderTrancheAvailable={!token.founder_tokens_claimed || token.founder_tokens_claimed === "0"}
                                    />
                                )}
                            />
                        </TabsContent>

                        <TabsContent value="all">
                            <TokenTable
                                tokens={allTokens}
                                emptyMessage="No tokens found."
                                isLoading={isFetchingTokens}
                                error={tokensError}
                                restBaseUrl={restBaseUrl}
                            />
                        </TabsContent>

                        <TabsContent value="trade">
                            <TradeForm
                                tradeForm={tradeForm}
                                tradeMode={tradeMode}
                                onTradeModeChange={handleTradeModeChange}
                                onInputChange={handleTradeInputChange}
                                onSubmit={handleBuyTokens}
                                isSubmitting={isSubmitting}
                                walletAddress={walletAddress}
                                availableBuyTokens={availableBuyTokens}
                                availableSellTokens={availableSellTokens}
                                tradePreview={tradePreview}
                            />
                        </TabsContent>

                        <TabsContent value="leverage">
                            <LeverageForm
                                walletAddress={walletAddress}
                                restEndpoint={network.restEndpoint}
                                registry={registry}
                                signAndBroadcast={signAndBroadcast}
                            />
                        </TabsContent>
                    </Tabs>
                </header>
            </div>
        </main>
    );
}

export default App;
