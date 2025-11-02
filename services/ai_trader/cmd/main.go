package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/bot"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/authz"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/feegrant"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/config"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/llm"
	md "github.com/osmosis-labs/osmosis/v30/services/ai_trader/marketdata"
)

const defaultSQLiteDSN = "file:ai_trader_md.db?cache=shared&_journal_mode=WAL"

func main() {
	root := &cobra.Command{Use: "ai-trader"}
	root.AddCommand(
		newRegisterCmd(),
		newGrantCmd(),
		newStartCmd(),
	)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newRegisterCmd() *cobra.Command {
	var (
		cfgPath  string
		dbDSN    string
		botName  string
		email    string
		userName string
	)
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new AI trader user and default bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(email) == "" {
				return fmt.Errorf("email required")
			}
			cfg, err := resolveConfig(cfgPath)
			if err != nil {
				return err
			}
			if botName == "" {
				botName = cfg.Bot.Name
			}
			if dbDSN == "" {
				dbDSN = defaultSQLiteDSN
			}
			repo, err := md.NewSQLiteRepository(dbDSN)
			if err != nil {
				return fmt.Errorf("open repo: %w", err)
			}
			defer repo.Close()
			uid, err := repo.CreateUser(email, strings.TrimSpace(userName))
			if err != nil {
				return fmt.Errorf("create user: %w", err)
			}
			cfgJSON, _ := md.DefaultPerspectiveConfig().Marshal()
			botID, err := repo.CreateBot(uid, botName, cfgJSON)
			if err != nil {
				return fmt.Errorf("create bot: %w", err)
			}
			apiKey, err := repo.CreateAPIKey(uid, md.ScopeUserAdmin, nil)
			if err != nil {
				return fmt.Errorf("create api key: %w", err)
			}
			output := map[string]any{
				"user_id": uid,
				"bot_id":  botID,
				"api_key": apiKey,
			}
			b, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	cmd.Flags().StringVar(&cfgPath, "config", "", "path to ai-trader config file (optional)")
	cmd.Flags().StringVar(&dbDSN, "db", "", fmt.Sprintf("sqlite DSN (default %s)", defaultSQLiteDSN))
	cmd.Flags().StringVar(&botName, "bot", "", "bot name (defaults to config bot name)")
	cmd.Flags().StringVar(&email, "email", "", "user email")
	cmd.Flags().StringVar(&userName, "name", "", "user display name")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func newGrantCmd() *cobra.Command {
	var (
		cfgPath        string
		nodeURL        string
		granter        string
		grantee        string
		expiration     string
		includeBonding bool
		withFeeGrant   bool
		feeSpend       string
		feeAllowed     []string
		feeExpiration  string
		timeout        time.Duration
	)
	cmd := &cobra.Command{
		Use:   "grant",
		Short: "Issue authz (and optional feegrant) for the configured bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := resolveConfig(cfgPath)
			if err != nil {
				return err
			}
			if timeout <= 0 {
				timeout = 45 * time.Second
			}
			if strings.TrimSpace(nodeURL) == "" {
				nodeURL = cfg.Bot.NodeURL
			}
			nodeURL = normalizeGRPCAddress(nodeURL)
			if nodeURL == "" {
				return fmt.Errorf("node URL is required (use --grpc or set bot.node_url)")
			}
			if strings.TrimSpace(granter) == "" {
				granter = cfg.Bot.GranterAddress
			}
			if strings.TrimSpace(grantee) == "" {
				grantee = cfg.Bot.GranteeAddress
			}
			authzExpiration, err := parseExpirationWithDefault(expiration, time.Time{}, 30*24*time.Hour)
			if err != nil {
				return fmt.Errorf("authz expiration: %w", err)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			authzClient, err := authz.NewClient(nodeURL)
			if err != nil {
				cancel()
				return fmt.Errorf("authz client: %w", err)
			}
			granted, err := authzClient.GrantTradingAuthorizations(ctx, granter, grantee, authzExpiration, includeBonding)
			cancel()
			authzClient.Close() //nolint:errcheck // best effort close
			if err != nil {
				return fmt.Errorf("grant authz: %w", err)
			}

			result := map[string]any{
				"node": nodeURL,
				"authz": map[string]any{
					"granter":    granter,
					"grantee":    grantee,
					"messages":   granted,
					"expiration": authzExpiration.UTC().Format(time.RFC3339),
				},
			}

			if withFeeGrant {
				if len(feeAllowed) == 0 {
					feeAllowed = append(feeAllowed, granted...)
				}
				var coins sdk.Coins
				if strings.TrimSpace(feeSpend) == "" {
					coins = sdk.NewCoins()
				} else {
					parsed, err := sdk.ParseCoinsNormalized(feeSpend)
					if err != nil {
						return fmt.Errorf("fee spend limit: %w", err)
					}
					coins = parsed
				}
				feeExp, err := parseExpirationWithDefault(feeExpiration, authzExpiration, 30*24*time.Hour)
				if err != nil {
					return fmt.Errorf("feegrant expiration: %w", err)
				}

				ctxFee, cancelFee := context.WithTimeout(cmd.Context(), timeout)
				fgClient, err := feegrant.NewClient(nodeURL)
				if err != nil {
					cancelFee()
					return fmt.Errorf("feegrant client: %w", err)
				}
				_, err = fgClient.GrantBasicAllowance(ctxFee, granter, grantee, coins, feeExp, feeAllowed)
				cancelFee()
				fgClient.Close() //nolint:errcheck
				if err != nil {
					return fmt.Errorf("feegrant: %w", err)
				}
				result["feegrant"] = map[string]any{
					"granter":          granter,
					"grantee":          grantee,
					"spend_limit":      coins.String(),
					"allowed_messages": feeAllowed,
					"expiration":       feeExp.UTC().Format(time.RFC3339),
				}
			}

			b, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	cmd.Flags().StringVar(&cfgPath, "config", "", "path to ai-trader config file")
	cmd.Flags().StringVar(&nodeURL, "grpc", "", "gRPC endpoint (defaults to bot.node_url)")
	cmd.Flags().StringVar(&granter, "granter", "", "override granter address")
	cmd.Flags().StringVar(&grantee, "grantee", "", "override grantee address")
	cmd.Flags().StringVar(&expiration, "expiration", "", "authz expiration (duration or RFC3339, default 720h)")
	cmd.Flags().BoolVar(&includeBonding, "bonding", false, "include bonding curve message authorizations")
	cmd.Flags().BoolVar(&withFeeGrant, "feegrant", false, "also issue a feegrant allowance")
	cmd.Flags().StringVar(&feeSpend, "fee-spend", "", "feegrant spend limit (e.g. 500000factory/test/ndollar)")
	cmd.Flags().StringSliceVar(&feeAllowed, "fee-allowed", nil, "restrict feegrant to message type URLs (default same as authz)")
	cmd.Flags().StringVar(&feeExpiration, "fee-expiration", "", "feegrant expiration (duration or RFC3339, defaults to authz expiration)")
	cmd.Flags().DurationVar(&timeout, "timeout", 45*time.Second, "RPC timeout per request")
	return cmd
}

func newStartCmd() *cobra.Command {
	var (
		cfgPath       string
		dbDSN         string
		grpcNode      string
		source        string
		symbols       []string
		disableREST   bool
		fetchInterval time.Duration
	)
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the AI trader bot loop with market data pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := resolveConfig(cfgPath)
			if err != nil {
				return err
			}
			if strings.TrimSpace(grpcNode) != "" {
				cfg.Bot.NodeURL = grpcNode
			}
			cfg.Bot.NodeURL = normalizeGRPCAddress(cfg.Bot.NodeURL)
			if cfg.Bot.NodeURL == "" {
				return fmt.Errorf("bot.node_url is required (or provide --grpc)")
			}
			if dbDSN == "" {
				dbDSN = defaultSQLiteDSN
			}
			repo, err := md.NewSQLiteRepository(dbDSN)
			if err != nil {
				return fmt.Errorf("open repo: %w", err)
			}
			defer repo.Close()

			sourceClean := strings.TrimSpace(source)
			if sourceClean == "" {
				sourceClean = "yahoo_http"
			}
			service := md.NewServiceWithSource(sourceClean).WithRepository(repo)

			if len(symbols) == 0 {
				symbols = append([]string(nil), cfg.Limits.AllowedSymbols...)
			}
			symbols = sanitizeSymbols(symbols)
			if len(symbols) == 0 {
				return fmt.Errorf("no symbols configured; set limits.allowed_symbols or pass --symbols")
			}

			if fetchInterval <= 0 {
				if cfgInterval, err := cfg.DataSources.PriceUpdateInterval.ParseDuration(); err == nil && cfgInterval > 0 {
					fetchInterval = cfgInterval
				} else {
					fetchInterval = 15 * time.Second
				}
			}

			timeframes := []md.Timeframe{md.TF1m, md.TF5m, md.TF1h, md.TF1d}
			scheduler := md.NewScheduler(service, symbols, timeframes, fetchInterval)
			scheduler.Start()
			defer scheduler.Stop()
			log.Printf("scheduler started (%s) for symbols=%v", fetchInterval, symbols)

			provider := llm.NewGroq("", "", 12*time.Second)
			runner, err := bot.NewRunner(cfg, service, provider)
			if err != nil {
				return fmt.Errorf("create runner: %w", err)
			}
			defer runner.Close()
			runner.SetLogger(log.Printf)

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			group, ctx := errgroup.WithContext(ctx)
			group.Go(func() error {
				if err := runner.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
					return err
				}
				return nil
			})

			if !disableREST {
				_ = os.Setenv("AI_TRADER_NODE_GRPC", cfg.Bot.NodeURL)
				group.Go(func() error {
					if err := serveREST(ctx, cfg, service); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
						return err
					}
					return nil
				})
			}

			log.Printf("runner started for bot=%s granter=%s grantee=%s", cfg.Bot.Name, cfg.Bot.GranterAddress, cfg.Bot.GranteeAddress)

			if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&cfgPath, "config", "", "path to ai-trader config file")
	cmd.Flags().StringVar(&dbDSN, "db", "", fmt.Sprintf("sqlite DSN (default %s)", defaultSQLiteDSN))
	cmd.Flags().StringVar(&grpcNode, "grpc", "", "override gRPC endpoint for trading client")
	cmd.Flags().StringVar(&source, "source", "yahoo_http", "market data source (yahoo_http|yahoo)")
	cmd.Flags().StringSliceVar(&symbols, "symbols", nil, "symbols to track (defaults to limits.allowed_symbols)")
	cmd.Flags().BoolVar(&disableREST, "no-rest", false, "disable the embedded REST server")
	cmd.Flags().DurationVar(&fetchInterval, "fetch-interval", 0, "override scheduler fetch interval")
	return cmd
}

func resolveConfig(path string) (*config.Config, error) {
	if strings.TrimSpace(path) == "" {
		return config.DefaultConfig(), nil
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}

func normalizeGRPCAddress(addr string) string {
	addr = strings.TrimSpace(addr)
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	return strings.TrimSpace(addr)
}

func parseExpirationWithDefault(input string, defaultTime time.Time, defaultDur time.Duration) (time.Time, error) {
	if strings.TrimSpace(input) == "" {
		if !defaultTime.IsZero() {
			return defaultTime.UTC(), nil
		}
		return time.Now().UTC().Add(defaultDur), nil
	}
	if d, err := time.ParseDuration(input); err == nil {
		if d <= 0 {
			return time.Time{}, fmt.Errorf("expiration duration must be positive")
		}
		return time.Now().UTC().Add(d), nil
	}
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("invalid expiration %q", input)
}

func serveREST(ctx context.Context, cfg *config.Config, svc *md.Service) error {
	mux := http.NewServeMux()
	rateInterval, err := cfg.API.RateInterval.ParseDuration()
	if err != nil || rateInterval <= 0 {
		rateInterval = time.Minute
	}
	limiter := md.NewRateLimiter(cfg.API.RateLimit, rateInterval)
	rest := md.NewREST(svc, md.WithRateLimiter(limiter), md.WithCORS(cfg.API.CORSOrigins))
	rest.Register(mux)

	server := &http.Server{
		Addr:    cfg.API.Bind,
		Handler: mux,
	}
	errCh := make(chan error, 1)
	go func() {
		cert := strings.TrimSpace(cfg.API.TLSCertPath)
		key := strings.TrimSpace(cfg.API.TLSKeyPath)
		var serveErr error
		if cert != "" && key != "" {
			log.Printf("REST listening with TLS on %s", cfg.API.Bind)
			serveErr = server.ListenAndServeTLS(cert, key)
		} else {
			log.Printf("REST listening on %s", cfg.API.Bind)
			serveErr = server.ListenAndServe()
		}
		errCh <- serveErr
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("REST shutdown error: %v", err)
		}
		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) || err == nil {
			return nil
		}
		return err
	case err := <-errCh:
		return err
	}
}

func sanitizeSymbols(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, sym := range in {
		s := strings.ToUpper(strings.TrimSpace(sym))
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
