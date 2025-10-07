import type { UserToken, UserTokensResponse } from '@/types';
import Decimal from "decimal.js";
import { useCallback, useEffect, useMemo, useState } from 'react';

interface UseTokensProps {
    restBaseUrl: string;
    walletAddress: string;
}

export function useTokens({ restBaseUrl, walletAddress }: UseTokensProps) {
    const [allTokens, setAllTokens] = useState<UserToken[]>([]);
    const [myTokens, setMyTokens] = useState<UserToken[]>([]);
    const [isFetchingTokens, setIsFetchingTokens] = useState(false);
    const [tokensError, setTokensError] = useState<string | null>(null);
    const [priceCache, setPriceCache] = useState<Record<string, Decimal>>({});

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
            setPriceCache((prev) => {
                // Only update if the price has actually changed
                if (prev[denom] && prev[denom].equals(price)) {
                    return prev;
                }
                return { ...prev, [denom]: price };
            });
        } catch (priceError) {
            console.error("Price fetch error:", priceError);
            setPriceCache((prev) => {
                // Only update if we don't already have a zero price
                if (prev[denom] && prev[denom].equals(new Decimal(0))) {
                    return prev;
                }
                return { ...prev, [denom]: new Decimal(0) };
            });
        }
    }, [restBaseUrl]);

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

    return useMemo(() => ({
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
    }), [
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
    ]);
}
