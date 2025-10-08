package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
)

func DecToCoin(amount osmomath.Dec, denom string) (sdk.Coin, error) {
	if amount.IsNegative() {
		return sdk.Coin{}, fmt.Errorf("negative amount")
	}

	intAmount := amount.TruncateInt()
	if !intAmount.IsPositive() {
		return sdk.Coin{}, fmt.Errorf("zero amount")
	}

	return sdk.NewCoin(denom, intAmount), nil
}

func CoinToDec(coin sdk.Coin) osmomath.Dec {
	return osmomath.NewDecFromInt(coin.Amount)
}
