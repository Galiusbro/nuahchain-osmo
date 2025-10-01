# User Token Creator UI

A minimal React + CosmJS front-end that lets users craft and broadcast
`MsgCreateUserToken` transactions through Keplr without hand-writing JSON.

## Prerequisites

- Node.js 18+
- Keplr browser extension (the app will try to suggest the Nuahchain config if
  the extension supports `experimentalSuggestChain`).
- A reachable RPC endpoint for the target chain.

## Install

```bash
cd apps/usertoken-ui
npm install
```

## Develop

```bash
npm run dev
```

The Vite dev server starts on port 5173 by default.

## Build

```bash
npm run build
```

## Test

```bash
npm test
```

## Usage

1. Fill in chain ID, RPC+REST endpoints, and the currency metadata (defaults are set for `NUAH`/`unuah`).
2. Click **Connect Keplr** and approve the requested permissions.
3. Enter the token metadata (subdenom, name, symbol, decimals, and optional
   memo).
4. Submit the form to sign and broadcast the transaction through Keplr.
5. Use the header tabs to switch between creation, your tokens (with founder
   tranche purchase at the special price), and the full list of user tokens on
   the network. The **Trade tokens** tab lets you buy or sell using the
   dropdown of available denoms so you always target an existing token.

The status panel shows the transaction hash, height, gas usage, and raw log so
operators can verify inclusion or debug failures.
