package types

import (
	"github.com/osmosis-labs/osmosis/osmomath"
)

// ConsensusMinFee is a governance set parameter from prop 354 (https://www.mintscan.io/osmosis/proposals/354)
// It was intended to be .0025 uosmo / gas
// In v30, we set it to 0.01 uosmo / gas
// Modified for nuahchain to 0.0001 unuah / gas (0.0001 unuah per gas unit) for 1 cent max fee
// var ConsensusMinFee osmomath.Dec = osmomath.ZeroDec()
var ConsensusMinFee osmomath.Dec = osmomath.NewDecWithPrec(1, 4)
