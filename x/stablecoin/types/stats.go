package types

import sdkmath "cosmossdk.io/math"

// NewStats constructs a Stats object from integer values.
func NewStats(totalMinted, totalBurned sdkmath.Int) Stats {
	outstanding := totalMinted.Sub(totalBurned)
	return Stats{
		TotalMinted: totalMinted.String(),
		TotalBurned: totalBurned.String(),
		Outstanding: outstanding.String(),
	}
}
