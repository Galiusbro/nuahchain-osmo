package freeaccount

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	freeaccountcli "github.com/osmosis-labs/osmosis/v30/x/freeaccount/client/cli"
	freeaccountkeeper "github.com/osmosis-labs/osmosis/v30/x/freeaccount/keeper"
	freeaccounttypes "github.com/osmosis-labs/osmosis/v30/x/freeaccount/types"
)

var (
	_ module.AppModuleBasic = AppModule{}
	_ module.HasServices    = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module used by the freeaccount module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the freeaccount module's name.
func (AppModuleBasic) Name() string { return freeaccounttypes.ModuleName }

// RegisterLegacyAminoCodec registers the freeaccount module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	freeaccounttypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (a AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	freeaccounttypes.RegisterInterfaces(reg)
}

// DefaultGenesis returns default genesis state as raw bytes for the freeaccount
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(freeaccounttypes.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the freeaccount module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState freeaccounttypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", freeaccounttypes.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the freeaccount module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes here if needed
}

// GetTxCmd returns the transaction commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return freeaccountcli.GetTxCmd()
}

// GetQueryCmd returns the cli query commands for this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return freeaccountcli.GetQueryCmd(freeaccounttypes.StoreKey)
}

// AppModule implements the AppModule interface for the freeaccount module.
type AppModule struct {
	AppModuleBasic

	keeper freeaccountkeeper.Keeper
}

func NewAppModule(
	cdc codec.Codec,
	keeper freeaccountkeeper.Keeper,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	freeaccounttypes.RegisterMsgServer(cfg.MsgServer(), freeaccountkeeper.NewMsgServerImpl(am.keeper))
	freeaccounttypes.RegisterQueryServer(cfg.QueryServer(), freeaccountkeeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs the freeaccount module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState freeaccounttypes.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)
	am.keeper.InitGenesis(ctx, genState)
}

// ExportGenesis returns the freeaccount module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

//
// App Wiring Setup (simplified)
//

type ModuleInputs struct {
	Cdc           codec.Codec
	StoreService  store.KVStoreService
	Logger        log.Logger
	AccountKeeper freeaccounttypes.AccountKeeper
}

type ModuleOutputs struct {
	FreeAccountKeeper freeaccountkeeper.Keeper
	Module            appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	k := freeaccountkeeper.NewKeeper(
		in.Cdc,
		in.StoreService,
		in.Logger,
		authority.String(),
		in.AccountKeeper,
	)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{FreeAccountKeeper: k, Module: m}
}

// Module is the config object of the freeaccount module.
type Module struct {
	Authority string `mapstructure:"authority"`
}
