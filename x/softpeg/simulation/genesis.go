package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v30/x/softpeg/types"
)

// RandomizedGenState generates a random GenesisState for softpeg
func RandomizedGenState(simState *module.SimulationState) {
	var params types.Params
	simState.AppParams.GetOrGenerate(
		"softpeg", &params, simState.Rand,
		func(r *rand.Rand) { params = types.DefaultParams() },
	)

	softpegGenesis := types.GenesisState{
		Params: params,
	}

	bz, err := json.MarshalIndent(&softpegGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated softpeg parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&softpegGenesis)
}
