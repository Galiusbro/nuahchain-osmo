package simulation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v30/x/pegkeeper/types"
)

// RandomizedGenState generates a random GenesisState for pegkeeper
func RandomizedGenState(simState *module.SimulationState) {
	params := types.NewParams(
		"0.05",      // MaxDeviationThreshold
		"0.10",      // AdjustmentFactor
		time.Hour,   // MinAdjustmentInterval
		"0.02",      // MaxSupplyChangePerAdjustment
		"usdoracle", // OracleModule
		true,        // Enabled
		"nuah",      // TargetDenom
		"usd",       // ReferenceDenom
		"1.0",       // TargetPrice
	)

	pegkeeperGenesis := types.GenesisState{
		Params: params,
	}

	bz, err := json.MarshalIndent(&pegkeeperGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated pegkeeper parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&pegkeeperGenesis)
}

// ProposalMsgs returns a slice of random proposal messages for simulation
func ProposalMsgs() []simulation.WeightedProposalMsg {
	return []simulation.WeightedProposalMsg{}
}

// WeightedOperations returns the all the pegkeeper module operations with their respective weights.
func WeightedOperations(appParams simulation.AppParams, cdc interface{}, ak interface{}, bk interface{}, k interface{}) []simulation.WeightedOperation {
	return []simulation.WeightedOperation{}
}
