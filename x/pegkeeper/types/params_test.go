package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v30/x/pegkeeper/types"
)

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	require.NotNil(t, params)
	require.NotEmpty(t, params.MaxDeviationThreshold)
	require.NotEmpty(t, params.AdjustmentFactor)
	require.True(t, params.MinAdjustmentInterval > 0)
	require.NotEmpty(t, params.TargetDenom)
	require.NotEmpty(t, params.ReferenceDenom)
	require.NotEmpty(t, params.TargetPrice)
}

func TestParamsValidate(t *testing.T) {
	tests := []struct {
		name   string
		params types.Params
		valid  bool
	}{
		{
			name:   "default params",
			params: types.DefaultParams(),
			valid:  true,
		},
		{
			name: "empty max deviation threshold",
			params: types.Params{
				MaxDeviationThreshold:        "",
				AdjustmentFactor:             "0.01",
				MinAdjustmentInterval:        3600,
				MaxSupplyChangePerAdjustment: "0.05",
				OracleModule:                 "usdoracle",
				Enabled:                      true,
				TargetDenom:                  "nuah",
				ReferenceDenom:               "usd",
				TargetPrice:                  "1.0",
			},
			valid: false,
		},
		{
			name: "empty adjustment factor",
			params: types.Params{
				MaxDeviationThreshold:        "0.01",
				AdjustmentFactor:             "",
				MinAdjustmentInterval:        3600,
				MaxSupplyChangePerAdjustment: "0.05",
				OracleModule:                 "usdoracle",
				Enabled:                      true,
				TargetDenom:                  "nuah",
				ReferenceDenom:               "usd",
				TargetPrice:                  "1.0",
			},
			valid: false,
		},
		{
			name: "zero min adjustment interval",
			params: types.Params{
				MaxDeviationThreshold:        "0.01",
				AdjustmentFactor:             "0.01",
				MinAdjustmentInterval:        0,
				MaxSupplyChangePerAdjustment: "0.05",
				OracleModule:                 "usdoracle",
				Enabled:                      true,
				TargetDenom:                  "nuah",
				ReferenceDenom:               "usd",
				TargetPrice:                  "1.0",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}