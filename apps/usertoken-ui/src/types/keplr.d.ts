import type { OfflineSigner, OfflineDirectSigner } from "@cosmjs/proto-signing";

interface SuggestChainCurrency {
  coinDenom: string;
  coinMinimalDenom: string;
  coinDecimals: number;
  gasPriceStep?: {
    low: number;
    average: number;
    high: number;
  };
}

interface SuggestChainParams {
  chainId: string;
  chainName: string;
  rpc: string;
  rest: string;
  bip44: {
    coinType: number;
  };
  bech32Config: Record<string, string>;
  currencies: SuggestChainCurrency[];
  feeCurrencies: SuggestChainCurrency[];
  stakeCurrency: SuggestChainCurrency;
  gasPriceStep?: {
    low: number;
    average: number;
    high: number;
  };
  features?: string[];
}

interface Keplr {
  enable: (chainId: string) => Promise<void>;
  getOfflineSignerAuto: (chainId: string) => Promise<OfflineSigner | OfflineDirectSigner>;
  getOfflineSigner: (chainId: string) => OfflineSigner;
  experimentalSuggestChain?: (chainInfo: SuggestChainParams) => Promise<void>;
  suggestToken?: (chainId: string, contractAddress: string, viewingKey?: string) => Promise<void>;
}

declare global {
  interface Window {
    keplr?: Keplr;
  }
}

export {};
