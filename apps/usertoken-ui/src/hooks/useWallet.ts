import type { KeplrCurrency, KeplrCurrencyWithGas, NetworkConfigState } from '@/types';
import { Registry } from "@cosmjs/proto-signing";
import { SigningStargateClient, defaultRegistryTypes } from "@cosmjs/stargate";
import { useCallback, useMemo, useState } from 'react';
import {
    MSG_ADD_COLLATERAL_TYPE_URL,
    MSG_CLOSE_POSITION_TYPE_URL,
    MSG_LIQUIDATE_POSITION_TYPE_URL,
    MSG_OPEN_POSITION_TYPE_URL,
    MSG_PROVIDE_LIQUIDITY_TYPE_URL,
    MSG_REMOVE_COLLATERAL_TYPE_URL
} from "../codec/leverage";
import {
    MsgAddCollateral,
    MsgClosePosition,
    MsgLiquidatePosition,
    MsgOpenPosition,
    MsgProvideLiquidity,
    MsgRemoveCollateral
} from "../codec/leverage_proto";
import {
    MSG_BUY_FOUNDER_TOKENS_TYPE_URL,
    MSG_BUY_TOKENS_TYPE_URL,
    MSG_CREATE_USER_TOKEN_TYPE_URL,
    MSG_SELL_TOKENS_TYPE_URL,
    MsgBuyFounderTokens,
    MsgBuyTokens,
    MsgCreateUserToken,
    MsgSellTokens
} from "../codec/usertoken";

const buildBech32Config = (prefix: string) => ({
    bech32PrefixAccAddr: prefix,
    bech32PrefixAccPub: `${prefix}pub`,
    bech32PrefixValAddr: `${prefix}valoper`,
    bech32PrefixValPub: `${prefix}valoperpub`,
    bech32PrefixConsAddr: `${prefix}valcons`,
    bech32PrefixConsPub: `${prefix}valconspub`
});

export function useWallet(network: NetworkConfigState) {
    const [client, setClient] = useState<SigningStargateClient | null>(null);
    const [walletAddress, setWalletAddress] = useState<string>("");
    const [isConnecting, setIsConnecting] = useState(false);

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
        [baseCurrency, network.bech32Prefix, network.chainId, network.chainName, network.rpcEndpoint, network.restEndpoint, network.gasPriceLow, network.gasPriceAverage, network.gasPriceHigh]
    );

    const connectWallet = useCallback(async () => {
        try {
            setIsConnecting(true);

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

            return {
                client: signingClient,
                address: accounts[0].address
            };
        } catch (error) {
            setClient(null);
            setWalletAddress("");
            throw error;
        } finally {
            setIsConnecting(false);
        }
    }, [network, registry]);

    return {
        client,
        walletAddress,
        isConnecting,
        connectWallet,
        suggestChain,
        baseCurrency
    };
}
