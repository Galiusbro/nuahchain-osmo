import type { NetworkConfigState } from '@/types';
import { useCallback, useState } from 'react';

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

export function useNetwork() {
    const [network, setNetwork] = useState<NetworkConfigState>(defaultNetwork);

    const handleNetworkChange = useCallback((field: keyof NetworkConfigState) =>
        (event: React.ChangeEvent<HTMLInputElement>) => {
            const { value } = event.target;
            setNetwork((current) => ({ ...current, [field]: value }));
        }, []);

    return {
        network,
        setNetwork,
        handleNetworkChange
    };
}
