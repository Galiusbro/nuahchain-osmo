import type { TradeFormState, TradeMode, TradePreviewState, UserToken } from '@/types';
import Decimal from "decimal.js";
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

const MICRO_FACTOR = new Decimal(1_000_000);

interface UseTradingProps {
    allTokens: UserToken[];
    myTokens: UserToken[];
    priceCache: Record<string, Decimal>;
    fetchPrice: (denom: string) => Promise<void>;
    fetchTokenEstimate: (denom: string, paymentAmount: string) => Promise<Decimal | null>;
}

export function useTrading({
    allTokens,
    myTokens,
    priceCache,
    fetchPrice,
    fetchTokenEstimate
}: UseTradingProps) {
    const [tradeForm, setTradeForm] = useState<TradeFormState>({
        denom: "",
        paymentAmount: "",
        paymentDenom: "unuah",
        tokenAmount: "",
        minTokens: "",
        minPrice: ""
    });
    const [tradeMode, setTradeMode] = useState<TradeMode>("buy");
    const [tradePreview, setTradePreview] = useState<TradePreviewState>({});

    // Use ref to stabilize fetchTokenEstimate function
    const fetchTokenEstimateRef = useRef(fetchTokenEstimate);
    fetchTokenEstimateRef.current = fetchTokenEstimate;

    // Memoize price cache to prevent unnecessary re-renders
    const stablePriceCache = useMemo(() => priceCache, [Object.keys(priceCache).join(',')]);

    // Use ref to track previous values and prevent unnecessary updates
    const prevValuesRef = useRef({
        tradeMode: tradeMode,
        denom: tradeForm.denom,
        paymentAmount: tradeForm.paymentAmount,
        tokenAmount: tradeForm.tokenAmount,
        priceCacheKeys: Object.keys(priceCache).join(',')
    });

    const availableBuyTokens = useMemo(() =>
        allTokens.filter(token => isTokenWithProperDecimals(token)),
        [allTokens]
    );

    const availableSellTokens = useMemo(() =>
        myTokens.filter(token => isTokenWithProperDecimals(token)),
        [myTokens]
    );

    const handleTradeModeChange = useCallback((mode: TradeMode) => {
        setTradeMode(mode);
        setTradeForm(prev => ({
            ...prev,
            denom: mode === "buy"
                ? availableBuyTokens[0]?.denom ?? ""
                : availableSellTokens[0]?.denom ?? ""
        }));
    }, [availableBuyTokens, availableSellTokens]);

    const handleTradeInputChange = useCallback((field: keyof TradeFormState) =>
        (event: React.ChangeEvent<HTMLInputElement>) => {
            const { value } = event.target;
            setTradeForm((current) => ({ ...current, [field]: value }));
        }, []);

    // Update trade form when available tokens change
    useEffect(() => {
        const newDenom = tradeMode === "buy"
            ? availableBuyTokens[0]?.denom ?? ""
            : availableSellTokens[0]?.denom ?? "";

        if (newDenom && newDenom !== tradeForm.denom) {
            setTradeForm(prev => ({
                ...prev,
                denom: newDenom
            }));
        }
    }, [tradeMode, availableBuyTokens, availableSellTokens, tradeForm.denom]);

    // Fetch price when token changes
    useEffect(() => {
        if (tradeForm.denom) {
            void fetchPrice(tradeForm.denom);
        }
    }, [tradeForm.denom, fetchPrice]);

    // Update trade preview
    useEffect(() => {
        const currentValues = {
            tradeMode: tradeMode,
            denom: tradeForm.denom,
            paymentAmount: tradeForm.paymentAmount,
            tokenAmount: tradeForm.tokenAmount,
            priceCacheKeys: Object.keys(priceCache).join(',')
        };

        // Check if values have actually changed
        const hasChanged =
            prevValuesRef.current.tradeMode !== currentValues.tradeMode ||
            prevValuesRef.current.denom !== currentValues.denom ||
            prevValuesRef.current.paymentAmount !== currentValues.paymentAmount ||
            prevValuesRef.current.tokenAmount !== currentValues.tokenAmount ||
            prevValuesRef.current.priceCacheKeys !== currentValues.priceCacheKeys;

        if (!hasChanged) {
            return;
        }

        // Update previous values
        prevValuesRef.current = currentValues;

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

                    const tokensUnits = await fetchTokenEstimateRef.current(tradeForm.denom, tradeForm.paymentAmount);
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

                    const price = stablePriceCache[tradeForm.denom];
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
    }, [tradeMode, tradeForm.denom, tradeForm.paymentAmount, tradeForm.tokenAmount, stablePriceCache, priceCache]);

    return {
        tradeForm,
        tradeMode,
        tradePreview,
        availableBuyTokens,
        availableSellTokens,
        handleTradeModeChange,
        handleTradeInputChange,
        setTradeForm
    };
}

function isTokenWithProperDecimals(token: UserToken): boolean {
    try {
        const currentSupply = BigInt(token.current_supply || "0");
        const maxSupply = BigInt(token.max_supply || "0");
        return currentSupply > 1_000_000_000n || maxSupply > 1_000_000_000n;
    } catch {
        return false;
    }
}
