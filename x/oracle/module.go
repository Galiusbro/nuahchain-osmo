package oracle

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/client/cli"
	"github.com/osmosis-labs/osmosis/v30/x/oracle/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}

	_ appmodule.AppModule        = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
)

// AppModuleBasic implements the stateless module methods.
type AppModuleBasic struct{}

// NewAppModuleBasic constructs a new AppModuleBasic instance.
func NewAppModuleBasic(_ codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{}
}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bz) == 0 {
		return nil
	}

	var state types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &state); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return state.Validate()
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	_ = types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements the full module interface.
type AppModule struct {
	AppModuleBasic
	keeper    keeper.Keeper
	apiKeeper *keeper.APIKeeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	// Create APIKeeper using the base keeper's properties
	apiKeeper := keeper.NewAPIKeeper(cdc.(codec.BinaryCodec), k.StoreKey(), k.Authority())

	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		apiKeeper:      apiKeeper,
	}
}

func (AppModule) IsAppModule()        {}
func (AppModule) IsOnePerModuleType() {}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	if len(data) == 0 {
		return
	}

	var genState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genState)
	if err := genState.Validate(); err != nil {
		panic(err)
	}
	for _, price := range genState.Prices {
		am.keeper.SetPrice(ctx, price)
	}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state := &types.GenesisState{
		Prices: am.keeper.GetAllPrices(ctx),
	}
	return cdc.MustMarshalJSON(state)
}

func (AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) BeginBlock(ctx context.Context) error {
	// Update prices every 5 minutes (300 blocks assuming 1 second block time)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()

	// Update prices every 300 blocks (5 minutes)
	if blockHeight%300 == 0 {
		am.updatePrices(sdkCtx)
	}

	return nil
}

// updatePrices updates prices for default symbols
func (am AppModule) updatePrices(ctx sdk.Context) {
	if am.apiKeeper == nil {
		return
	}

	// Get default symbols
	symbols := am.apiKeeper.GetDefaultSymbols()

	// Update prices for each category
	for category, symbolList := range symbols {
		am.apiKeeper.Logger(ctx).Info("Updating prices", "category", category, "count", len(symbolList))

		// Update first few symbols from each category to avoid rate limits
		maxSymbols := 3
		if len(symbolList) < maxSymbols {
			maxSymbols = len(symbolList)
		}

		for i := 0; i < maxSymbols; i++ {
			symbol := symbolList[i]
			err := am.apiKeeper.UpdatePriceFromAPI(ctx, symbol)
			if err != nil {
				am.apiKeeper.Logger(ctx).Error("Failed to update price", "symbol", symbol, "error", err)
			} else {
				am.apiKeeper.Logger(ctx).Info("Updated price", "symbol", symbol)
			}
		}
	}
}

func (am AppModule) EndBlock(context.Context) error { return nil }
