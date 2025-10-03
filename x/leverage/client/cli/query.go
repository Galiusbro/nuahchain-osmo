package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group leverage queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryPosition())
	cmd.AddCommand(CmdQueryPositions())
	cmd.AddCommand(CmdQueryPositionsByTrader())
	cmd.AddCommand(CmdQueryLiquidationPrice())
	cmd.AddCommand(CmdQueryEstimatePosition())
	cmd.AddCommand(CmdQueryTokenPrice())

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "shows the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryPosition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "position [position-id]",
		Short: "Query a specific position by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			positionID := args[0]

			res, err := queryClient.Position(context.Background(), &types.QueryPositionRequest{
				PositionId: positionID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryPositions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "positions",
		Short: "Query all positions with optional filters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			status, _ := cmd.Flags().GetString("status")
			tokenDenom, _ := cmd.Flags().GetString("token-denom")

			var positionStatus types.PositionStatus
			switch status {
			case "open":
				positionStatus = types.PositionStatusOpen
			case "closed":
				positionStatus = types.PositionStatusClosed
			case "liquidated":
				positionStatus = types.PositionStatusLiquidated
			default:
				positionStatus = types.PositionStatusUnspecified
			}

			res, err := queryClient.Positions(context.Background(), &types.QueryPositionsRequest{
				Pagination: pageReq,
				Status:     positionStatus,
				TokenDenom: tokenDenom,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by position status (open, closed, liquidated)")
	cmd.Flags().String("token-denom", "", "Filter by token denomination")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)

	return cmd
}

func CmdQueryPositionsByTrader() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "positions-by-trader [trader-address]",
		Short: "Query all positions for a specific trader",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			trader := args[0]
			status, _ := cmd.Flags().GetString("status")

			var positionStatus types.PositionStatus
			switch status {
			case "open":
				positionStatus = types.PositionStatusOpen
			case "closed":
				positionStatus = types.PositionStatusClosed
			case "liquidated":
				positionStatus = types.PositionStatusLiquidated
			default:
				positionStatus = types.PositionStatusUnspecified
			}

			res, err := queryClient.PositionsByTrader(context.Background(), &types.QueryPositionsByTraderRequest{
				Trader:     trader,
				Pagination: pageReq,
				Status:     positionStatus,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by position status (open, closed, liquidated)")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)

	return cmd
}

func CmdQueryLiquidationPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidation-price [collateral-amount] [position-size] [entry-price] [side]",
		Short: "Calculate liquidation price for given parameters",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			collateralAmount, ok := math.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("invalid collateral amount: %s", args[0])
			}

			positionSize, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid position size: %s", args[1])
			}

			entryPrice, err := math.LegacyNewDecFromStr(args[2])
			if err != nil {
				return fmt.Errorf("invalid entry price: %w", err)
			}

			var side types.PositionSide
			switch args[3] {
			case "long", "LONG":
				side = types.PositionSideLong
			case "short", "SHORT":
				side = types.PositionSideShort
			default:
				return fmt.Errorf("invalid side: %s (must be 'long' or 'short')", args[3])
			}

			res, err := queryClient.LiquidationPrice(context.Background(), &types.QueryLiquidationPriceRequest{
				CollateralAmount: collateralAmount,
				PositionSize:     positionSize,
				EntryPrice:       entryPrice,
				Side:             side,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryEstimatePosition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate-position [token-denom] [collateral-amount] [leverage] [side]",
		Short: "Estimate the outcome of opening a position",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			tokenDenom := args[0]

			collateralAmount, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid collateral amount: %s", args[1])
			}

			leverage, err := math.LegacyNewDecFromStr(args[2])
			if err != nil {
				return fmt.Errorf("invalid leverage: %w", err)
			}

			var side types.PositionSide
			switch args[3] {
			case "long", "LONG":
				side = types.PositionSideLong
			case "short", "SHORT":
				side = types.PositionSideShort
			default:
				return fmt.Errorf("invalid side: %s (must be 'long' or 'short')", args[3])
			}

			res, err := queryClient.EstimatePosition(context.Background(), &types.QueryEstimatePositionRequest{
				TokenDenom:       tokenDenom,
				CollateralAmount: collateralAmount,
				Leverage:         leverage,
				Side:             side,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryTokenPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-price [denom]",
		Short: "Query the current price of a token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			denom := args[0]

			res, err := queryClient.TokenPrice(context.Background(), &types.QueryTokenPriceRequest{
				Denom: denom,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
