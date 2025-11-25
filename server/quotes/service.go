package quotes

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	bondingcurvetypes "github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
	exchangetypes "github.com/osmosis-labs/osmosis/v30/x/exchange/types"
	"github.com/osmosis-labs/osmosis/osmomath"
)

// Service handles quote calculations for trading and swapping
type Service struct {
	blockchainCli *blockchain.Client
}

// NewService creates a new quotes service
func NewService(blockchainCli *blockchain.Client) *Service {
	return &Service{
		blockchainCli: blockchainCli,
	}
}

// TradeQuoteRequest represents a request for trade quote
type TradeQuoteRequest struct {
	Denom        string `json:"denom"`         // Token denom (e.g., "factory/.../token")
	Operation    string `json:"operation"`      // "buy" or "sell"
	Amount       string `json:"amount"`         // Payment amount (buy) or token amount (sell)
	PaymentDenom string `json:"payment_denom"` // Payment currency (e.g., "unuah", "undollar")
}

// TradeQuoteResponse represents a trade quote response
type TradeQuoteResponse struct {
	Denom         string `json:"denom"`
	Operation     string `json:"operation"`      // "buy" or "sell"
	InputAmount   string `json:"input_amount"`   // Amount user pays (buy) or sells (sell)
	InputDenom    string `json:"input_denom"`   // Input currency
	OutputAmount  string `json:"output_amount"`  // Amount user receives
	OutputDenom   string `json:"output_denom"`  // Output currency
	Price         string `json:"price"`          // Current price per token
	PriceImpact   string `json:"price_impact"`   // Estimated price impact (%)
	Fee           string `json:"fee,omitempty"`  // Protocol fee (if applicable)
	MinOutput     string `json:"min_output"`    // Minimum output (for slippage protection)
}

// SwapQuoteRequest represents a request for swap quote
type SwapQuoteRequest struct {
	TokenIn  string `json:"token_in"`  // Input token denom (e.g., "ueth", "ubtc")
	AmountIn string `json:"amount_in"` // Amount of input token
}

// SwapQuoteResponse represents a swap quote response
type SwapQuoteResponse struct {
	TokenIn     string `json:"token_in"`
	AmountIn    string `json:"amount_in"`
	TokenOut    string `json:"token_out"`    // Always "unuah" (N$)
	AmountOut   string `json:"amount_out"`   // Estimated unuah output
	ExchangeRate string `json:"exchange_rate"` // Exchange rate (tokens per unuah)
	Fee         string `json:"fee,omitempty"`  // Exchange fee (if applicable)
	MinOutput   string `json:"min_output"`    // Minimum output (for slippage protection)
}

// GetSupportedTokensResponse represents supported tokens for exchange
type GetSupportedTokensResponse struct {
	SupportedTokens []string `json:"supported_tokens"` // Tokens configured in exchange params
	AvailableRates  []string `json:"available_rates"`  // Tokens with active exchange rates
}

// GetTradeQuote calculates a quote for buying or selling tokens on bonding curve
func (s *Service) GetTradeQuote(ctx context.Context, req TradeQuoteRequest) (*TradeQuoteResponse, error) {
	if s.blockchainCli == nil {
		return nil, fmt.Errorf("blockchain client not configured")
	}

	// Get bonding curve params
	paramsReq := &bondingcurvetypes.QueryParamsRequest{}
	paramsResp, err := s.blockchainCli.BondingQueryClient.Params(ctx, paramsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get bonding curve params: %w", err)
	}
	params := paramsResp.Params

	// Get token list to find tokens sold (TokenSummary has tokens_sold field)
	// We'll use ListTokens and find our token, or use a default of 0
	tokensSold := osmomath.ZeroDec()
	listReq := &bondingcurvetypes.QueryListTokensRequest{Limit: 1000, Offset: 0}
	listResp, err := s.blockchainCli.BondingQueryClient.ListTokens(ctx, listReq)
	if err == nil {
		for _, token := range listResp.Tokens {
			if token.Denom == req.Denom {
				if token.TokensSold != "" {
					tokensSold = osmomath.MustNewDecFromStr(token.TokensSold)
				}
				break
			}
		}
	}

	// Set default payment denom
	paymentDenom := req.PaymentDenom
	if paymentDenom == "" {
		paymentDenom = params.QuoteDenom
		if paymentDenom == "" {
			paymentDenom = "unuah" // Default
		}
	}

	// Parse input amount
	inputAmount, err := osmomath.NewDecFromStr(req.Amount)
	if err != nil || !inputAmount.IsPositive() {
		return nil, fmt.Errorf("invalid amount: %s", req.Amount)
	}

	var outputAmount osmomath.Dec
	var outputDenom string
	var inputDenom string
	var currentPrice osmomath.Dec

	// Calculate current price
	currentPrice = bondingcurvetypes.CalculatePrice(tokensSold, params)

	if req.Operation == "buy" {
		// Calculate tokens out for given payment amount
		// Account for protocol fee
		netPayment := inputAmount
		if params.ProtocolFeeRateDec().IsPositive() {
			feeRate := params.ProtocolFeeRateDec()
			fee := inputAmount.Mul(feeRate)
			netPayment = inputAmount.Sub(fee)
		}

		outputAmount = bondingcurvetypes.IntegrateBuyAmount(tokensSold, netPayment, params)
		outputDenom = req.Denom
		inputDenom = paymentDenom

		// Calculate price impact (simplified: average price vs current price)
		// This is an approximation
		if !outputAmount.IsZero() {
			avgPrice := netPayment.Quo(outputAmount)
			if !currentPrice.IsZero() {
				priceImpact := avgPrice.Sub(currentPrice).Quo(currentPrice).Mul(osmomath.NewDec(100))
				// Return absolute value
				if priceImpact.IsNegative() {
					priceImpact = priceImpact.Neg()
				}
				// Cap at 100%
				if priceImpact.GT(osmomath.NewDec(100)) {
					priceImpact = osmomath.NewDec(100)
				}
			}
		}
	} else if req.Operation == "sell" {
		// Calculate payment out for given token amount
		outputAmount = bondingcurvetypes.IntegrateSellAmount(tokensSold, inputAmount, params)

		// Account for protocol fee
		if params.ProtocolFeeRateDec().IsPositive() {
			feeRate := params.ProtocolFeeRateDec()
			fee := outputAmount.Mul(feeRate)
			outputAmount = outputAmount.Sub(fee)
		}

		outputDenom = paymentDenom
		inputDenom = req.Denom

		// Calculate price impact (simplified)
		if !inputAmount.IsZero() {
			avgPrice := outputAmount.Quo(inputAmount)
			if !currentPrice.IsZero() {
				priceImpact := currentPrice.Sub(avgPrice).Quo(currentPrice).Mul(osmomath.NewDec(100))
				// Return absolute value
				if priceImpact.IsNegative() {
					priceImpact = priceImpact.Neg()
				}
				// Cap at 100%
				if priceImpact.GT(osmomath.NewDec(100)) {
					priceImpact = osmomath.NewDec(100)
				}
			}
		}
	} else {
		return nil, fmt.Errorf("invalid operation: %s (must be 'buy' or 'sell')", req.Operation)
	}

	if !outputAmount.IsPositive() {
		return nil, fmt.Errorf("calculated output amount is not positive")
	}

	// Calculate fee (if applicable)
	var feeStr string
	if params.ProtocolFeeRateDec().IsPositive() {
		if req.Operation == "buy" {
			fee := inputAmount.Mul(params.ProtocolFeeRateDec())
			feeStr = fee.String()
		} else {
			// For sell, fee is deducted from output
			netOutput := bondingcurvetypes.IntegrateSellAmount(tokensSold, inputAmount, params)
			fee := netOutput.Mul(params.ProtocolFeeRateDec())
			feeStr = fee.String()
		}
	}

	// Calculate min output (with 0.5% slippage tolerance by default)
	slippageTolerance := osmomath.MustNewDecFromStr("0.995") // 0.5% slippage
	minOutput := outputAmount.Mul(slippageTolerance)

	// Calculate price impact (simplified approximation)
	priceImpact := osmomath.ZeroDec()
	if req.Operation == "buy" && !outputAmount.IsZero() {
		avgPrice := inputAmount.Quo(outputAmount)
		if !currentPrice.IsZero() {
			impact := avgPrice.Sub(currentPrice).Quo(currentPrice).Mul(osmomath.NewDec(100))
			if impact.IsNegative() {
				impact = impact.Neg()
			}
			if impact.LTE(osmomath.NewDec(100)) {
				priceImpact = impact
			} else {
				priceImpact = osmomath.NewDec(100)
			}
		}
	} else if req.Operation == "sell" && !inputAmount.IsZero() {
		avgPrice := outputAmount.Quo(inputAmount)
		if !currentPrice.IsZero() {
			impact := currentPrice.Sub(avgPrice).Quo(currentPrice).Mul(osmomath.NewDec(100))
			if impact.IsNegative() {
				impact = impact.Neg()
			}
			if impact.LTE(osmomath.NewDec(100)) {
				priceImpact = impact
			} else {
				priceImpact = osmomath.NewDec(100)
			}
		}
	}

	return &TradeQuoteResponse{
		Denom:        req.Denom,
		Operation:    req.Operation,
		InputAmount:   inputAmount.String(),
		InputDenom:    inputDenom,
		OutputAmount: outputAmount.String(),
		OutputDenom:  outputDenom,
		Price:         currentPrice.String(),
		PriceImpact:   priceImpact.String(),
		Fee:           feeStr,
		MinOutput:     minOutput.String(),
	}, nil
}

// GetSwapQuote calculates a quote for swapping tokens via exchange module
func (s *Service) GetSwapQuote(ctx context.Context, req SwapQuoteRequest) (*SwapQuoteResponse, error) {
	if s.blockchainCli == nil {
		return nil, fmt.Errorf("blockchain client not configured")
	}

	// First, check if token is supported by getting exchange params
	paramsReq := &exchangetypes.QueryParamsRequest{}
	paramsResp, err := s.blockchainCli.ExchangeQueryClient.Params(ctx, paramsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange params: %w", err)
	}

	params := paramsResp.Params
	if !params.Enabled {
		return nil, fmt.Errorf("exchange module is disabled")
	}

	// Check if token is in supported tokens list
	isSupported := false
	for _, supportedToken := range params.SupportedTokens {
		if supportedToken == req.TokenIn {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return nil, fmt.Errorf("token %s is not supported. Supported tokens: %v", req.TokenIn, params.SupportedTokens)
	}

	// Get exchange rate for the token
	exchangeRateReq := &exchangetypes.QueryExchangeRateRequest{Denom: req.TokenIn}
	exchangeRateResp, err := s.blockchainCli.ExchangeQueryClient.ExchangeRate(ctx, exchangeRateReq)
	if err != nil {
		// Try to get all exchange rates to see what's available
		allRatesReq := &exchangetypes.QueryExchangeRatesRequest{}
		allRatesResp, allErr := s.blockchainCli.ExchangeQueryClient.ExchangeRates(ctx, allRatesReq)
		if allErr == nil && len(allRatesResp.ExchangeRates) > 0 {
			availableDenoms := make([]string, 0, len(allRatesResp.ExchangeRates))
			for _, rate := range allRatesResp.ExchangeRates {
				availableDenoms = append(availableDenoms, rate.Denom)
			}
			return nil, fmt.Errorf("exchange rate not found for token %s. Available rates: %v. Note: Exchange rates need to be updated via UpdateExchangeRate (requires oracle and TWAP data)", req.TokenIn, availableDenoms)
		}
		return nil, fmt.Errorf("failed to get exchange rate for %s: %w. Token is supported but exchange rate is not set. Exchange rates need to be updated via UpdateExchangeRate", req.TokenIn, err)
	}

	exchangeRate := exchangeRateResp.ExchangeRate
	if exchangeRate.Rate.IsZero() || exchangeRate.Rate.IsNegative() {
		return nil, fmt.Errorf("exchange rate not available or invalid for token: %s", req.TokenIn)
	}

	// Parse amounts
	amountIn, err := osmomath.NewDecFromStr(req.AmountIn)
	if err != nil || !amountIn.IsPositive() {
		return nil, fmt.Errorf("invalid amount_in: %s", req.AmountIn)
	}

	// ExchangeRate.Rate is LegacyDec, convert to osmomath.Dec via string
	// LegacyDec has String() method that returns the decimal string
	rateStr := exchangeRate.Rate.String()
	rate, err := osmomath.NewDecFromStr(rateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse exchange rate: %w", err)
	}

	// Calculate output: amount_in * rate = unuah_out
	// Rate represents USD per token, so amount_in (in token units) * rate = USD value
	// Then convert to unuah (assuming 1 USD = 1 unuah for exchange)
	amountOut := amountIn.Mul(rate)

	// Get exchange params to check for fees
	// Note: Exchange module might have fees, but for now we'll return the base calculation
	// TODO: Add fee calculation if exchange module has fees

	// Calculate min output (with 0.5% slippage tolerance)
	slippageTolerance := osmomath.MustNewDecFromStr("0.995")
	minOutput := amountOut.Mul(slippageTolerance)

	return &SwapQuoteResponse{
		TokenIn:      req.TokenIn,
		AmountIn:     req.AmountIn,
		TokenOut:     "unuah", // Exchange always outputs unuah
		AmountOut:    amountOut.String(),
		ExchangeRate: rate.String(),
		MinOutput:    minOutput.String(),
	}, nil
}

// GetSupportedTokens returns list of supported tokens and available exchange rates
func (s *Service) GetSupportedTokens(ctx context.Context) (*GetSupportedTokensResponse, error) {
	if s.blockchainCli == nil {
		return nil, fmt.Errorf("blockchain client not configured")
	}

	// Get exchange params
	paramsReq := &exchangetypes.QueryParamsRequest{}
	paramsResp, err := s.blockchainCli.ExchangeQueryClient.Params(ctx, paramsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange params: %w", err)
	}

	// Get all exchange rates
	ratesReq := &exchangetypes.QueryExchangeRatesRequest{}
	ratesResp, err := s.blockchainCli.ExchangeQueryClient.ExchangeRates(ctx, ratesReq)
	if err != nil {
		// If rates query fails, just return supported tokens
		return &GetSupportedTokensResponse{
			SupportedTokens: paramsResp.Params.SupportedTokens,
			AvailableRates:  []string{},
		}, nil
	}

	availableDenoms := make([]string, 0, len(ratesResp.ExchangeRates))
	for _, rate := range ratesResp.ExchangeRates {
		availableDenoms = append(availableDenoms, rate.Denom)
	}

	return &GetSupportedTokensResponse{
		SupportedTokens: paramsResp.Params.SupportedTokens,
		AvailableRates:  availableDenoms,
	}, nil
}
