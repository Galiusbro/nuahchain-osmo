import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

const connectWithSignerMock = vi.hoisted(() => vi.fn());
const fetchMock = vi.hoisted(() => vi.fn());

vi.mock("@cosmjs/stargate", () => ({
    SigningStargateClient: { connectWithSigner: connectWithSignerMock },
    assertIsDeliverTxSuccess: vi.fn(),
    defaultRegistryTypes: []
}));

describe("Keplr connection flow", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        connectWithSignerMock.mockReset();
        fetchMock.mockResolvedValue({
            ok: true,
            json: async () => ({ user_tokens: [], pagination: { next_key: null } })
        });
        global.fetch = fetchMock as unknown as typeof fetch;
    });

    afterEach(() => {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore - vitest does not know about keplr on the window object
        delete window.keplr;
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore - clean up mocked fetch
        delete global.fetch;
    });

    it("shows an error when Keplr is unavailable", async () => {
        render(<App />);

        const connectButton = screen.getByRole("button", { name: /connect keplr/i });
        fireEvent.click(connectButton);

        await screen.findByText(/Keplr extension not detected/i);
        expect(connectWithSignerMock).not.toHaveBeenCalled();
    });

    it("connects successfully when Keplr is present", async () => {
        const mockAccounts = [{ address: "nuah1abcdefgh" }];
        const getAccounts = vi.fn().mockResolvedValue(mockAccounts);
        const experimentalSuggestChain = vi.fn().mockResolvedValue(undefined);

        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore - augmenting window for the test environment
        window.keplr = {
            experimentalSuggestChain,
            enable: vi.fn().mockResolvedValue(undefined),
            getOfflineSignerAuto: vi.fn().mockResolvedValue({ getAccounts })
        };

        connectWithSignerMock.mockResolvedValue({ signAndBroadcast: vi.fn() });

        render(<App />);

        const connectButton = screen.getByRole("button", { name: /connect keplr/i });
        fireEvent.click(connectButton);

        await waitFor(() => {
            expect(experimentalSuggestChain).toHaveBeenCalledTimes(1);
        });

        await waitFor(() => {
            expect(connectWithSignerMock).toHaveBeenCalledTimes(1);
        });

        await screen.findByText(/Wallet connected: nuah1abcdefgh/i);

        const keplr = window.keplr!;
        expect(keplr.enable).toHaveBeenCalledWith("nuahchain");
        expect(keplr.getOfflineSignerAuto).toHaveBeenCalledWith("nuahchain");
        expect(experimentalSuggestChain).toHaveBeenCalledWith(
            expect.objectContaining({
                chainId: "nuahchain",
                stakeCurrency: expect.objectContaining({ coinMinimalDenom: "unuah" })
            })
        );
    });
});
