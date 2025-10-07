import { Button } from '@/components/ui/button';
import type { UserToken } from '@/types';

interface TokenActionsProps {
    token: UserToken;
    onAddToKeplr: (token: UserToken) => void;
    onBuyFounderTokens: (token: UserToken) => void;
    isAddingToKeplr: boolean;
    isBuyingFounder: boolean;
    isAlreadyAddedToKeplr: boolean;
    isFounderTrancheAvailable: boolean;
}

const FOUNDER_TRANCHE_AMOUNT_DISPLAY = "10,000,000";
const FOUNDER_TRANCHE_COST_DISPLAY = "500 NUAH";

export default function TokenActions({
    token,
    onAddToKeplr,
    onBuyFounderTokens,
    isAddingToKeplr,
    isBuyingFounder,
    isAlreadyAddedToKeplr,
    isFounderTrancheAvailable
}: TokenActionsProps) {
    return (
        <div className="flex gap-2 flex-wrap">
            <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => onAddToKeplr(token)}
                disabled={isAddingToKeplr}
                className="text-blue-600 border-blue-600 hover:bg-blue-50"
            >
                {isAddingToKeplr
                    ? "Adding..."
                    : isAlreadyAddedToKeplr
                        ? "Update in Keplr"
                        : "Add to Keplr"
                }
            </Button>

            {isFounderTrancheAvailable ? (
                <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => onBuyFounderTokens(token)}
                    disabled={isBuyingFounder}
                    className="text-green-600 border-green-600 hover:bg-green-50"
                >
                    {isBuyingFounder
                        ? "Processing..."
                        : `Buy ${FOUNDER_TRANCHE_AMOUNT_DISPLAY} (${FOUNDER_TRANCHE_COST_DISPLAY})`
                    }
                </Button>
            ) : (
                <span className="text-sm text-gray-500 self-center px-2">
                    Founder tranche claimed
                </span>
            )}
        </div>
    );
}


