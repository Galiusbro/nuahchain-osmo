package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v30/x/policy/client/cli"
	"github.com/osmosis-labs/osmosis/v30/x/policy/keeper"
	"github.com/osmosis-labs/osmosis/v30/x/policy/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}

	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}
)

// AppModuleBasic implements stateless module functionality.
type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

// NewAppModuleBasic creates a new AppModuleBasic instance.
func NewAppModuleBasic(cdc codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name returns the module name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers Amino codecs (noop).
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterInterfaces registers message and query interfaces.
func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns the default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis validates genesis data.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var state types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &state); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return state.Validate()
}

// RegisterGRPCGatewayRoutes registers gRPC query handlers.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetTxCmd returns CLI transaction commands.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns CLI query commands.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements the module.AppModule interface.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule constructs a new AppModule.
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

// IsAppModule marks the struct as an appmodule.AppModule implementation.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType ensures only one module per type in depinject contexts.
func (AppModule) IsOnePerModuleType() {}

// RegisterServices registers message and query servers.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// RegisterInvariants registers module invariants (noop).
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis initializes module state from genesis data.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genState)

	InitGenesis(ctx, am.keeper, &genState)
}

// ExportGenesis exports module state.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion returns the module consensus version.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes begin blocker logic (noop).
func (am AppModule) BeginBlock(ctx context.Context) error {
	return nil
}

// EndBlock executes end blocker logic (noop).
func (am AppModule) EndBlock(ctx context.Context) error {
	return nil
}
