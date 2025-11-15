package marketplace

import (
	"context"
	"fmt"
	"strings"

	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
	"github.com/osmosis-labs/osmosis/v30/server/tokens"
	bondingcurvetypes "github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

// Service handles marketplace operations for tokens
type Service struct {
	blockchainCli *blockchain.Client
	tokensRepo    *tokens.Repository
}

// NewService creates a new marketplace service
func NewService(blockchainCli *blockchain.Client, tokensRepo *tokens.Repository) *Service {
	return &Service{
		blockchainCli: blockchainCli,
		tokensRepo:    tokensRepo,
	}
}

// GetMarketplaceTokens gets list of all tokens from blockchain
func (s *Service) GetMarketplaceTokens(ctx context.Context, limit, offset uint64) ([]TokenMarketInfo, error) {
	if s.blockchainCli == nil {
		return nil, fmt.Errorf("blockchain client not configured")
	}

	// Query tokens from bondingcurve module
	req := &bondingcurvetypes.QueryListTokensRequest{
		Limit:  limit,
		Offset: offset,
	}
	if limit == 0 {
		req.Limit = 100 // Default limit
	}

	resp, err := s.blockchainCli.BondingQueryClient.ListTokens(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens from blockchain: %w", err)
	}

	// Convert to our format and enrich with metadata from DB
	result := make([]TokenMarketInfo, 0, len(resp.Tokens))
	denoms := make([]string, 0, len(resp.Tokens))

	for _, token := range resp.Tokens {
		denoms = append(denoms, token.Denom)
		result = append(result, TokenMarketInfo{
			Denom:          token.Denom,
			Name:           token.Name,
			Symbol:         token.Symbol,
			Creator:        token.Creator,
			CurrentPrice:   token.CurrentPrice,
			TokensSold:     token.TokensSold,
			CurveCompleted: token.CurveCompleted,
			Stats:          token.Stats,
		})
	}

	// Enrich with metadata from database
	if s.tokensRepo != nil && len(denoms) > 0 {
		metadataMap, err := s.tokensRepo.GetTokensByDenoms(denoms)
		if err == nil {
			for i := range result {
				if metadata, found := metadataMap[result[i].Denom]; found {
					// Override with DB metadata if available (more complete)
					if metadata.Image != nil {
						result[i].Image = *metadata.Image
					}
					if metadata.Description != nil {
						result[i].Description = *metadata.Description
					}
					result[i].Decimals = metadata.Decimals
				}
			}
		}
	}

	return result, nil
}

// SearchTokens searches tokens by name or symbol
func (s *Service) SearchTokens(ctx context.Context, query string, limit, offset uint64) ([]TokenMarketInfo, error) {
	if query == "" {
		return s.GetMarketplaceTokens(ctx, limit, offset)
	}

	// Get all tokens and filter by query
	allTokens, err := s.GetMarketplaceTokens(ctx, 1000, 0) // Get more tokens for search
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	filtered := make([]TokenMarketInfo, 0)

	for _, token := range allTokens {
		// Search in name, symbol, or denom
		if strings.Contains(strings.ToLower(token.Name), queryLower) ||
			strings.Contains(strings.ToLower(token.Symbol), queryLower) ||
			strings.Contains(strings.ToLower(token.Denom), queryLower) {
			filtered = append(filtered, token)
		}
	}

	// Apply pagination
	start := int(offset)
	end := start + int(limit)
	if start >= len(filtered) {
		return []TokenMarketInfo{}, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

// GetTokenDetails gets detailed information about a specific token
func (s *Service) GetTokenDetails(ctx context.Context, denom string) (*TokenDetails, error) {
	if s.blockchainCli == nil {
		return nil, fmt.Errorf("blockchain client not configured")
	}

	// Get token stats from blockchain
	statsReq := &bondingcurvetypes.QueryTokenStatsRequest{Denom: denom}
	statsResp, err := s.blockchainCli.BondingQueryClient.TokenStats(ctx, statsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get token stats: %w", err)
	}

	details := &TokenDetails{
		Denom:  denom,
		Stats:  statsResp.Stats,
	}

	// Get metadata from database
	if s.tokensRepo != nil {
		metadata, err := s.tokensRepo.GetTokenByDenom(denom)
		if err == nil && metadata != nil {
			details.Name = metadata.Name
			details.Symbol = metadata.Symbol
			if metadata.Image != nil {
				details.Image = *metadata.Image
			}
			if metadata.Description != nil {
				details.Description = *metadata.Description
			}
			details.Creator = metadata.CreatorAddress
			details.Decimals = metadata.Decimals
		}
	}

	return details, nil
}

// TokenMarketInfo represents token information for marketplace listing
type TokenMarketInfo struct {
	Denom          string                        `json:"denom"`
	Name           string                        `json:"name"`
	Symbol         string                        `json:"symbol"`
	Image          string                        `json:"image,omitempty"`
	Description    string                        `json:"description,omitempty"`
	Creator        string                        `json:"creator"`
	CurrentPrice   string                        `json:"current_price,omitempty"`
	TokensSold     string                        `json:"tokens_sold,omitempty"`
	CurveCompleted bool                          `json:"curve_completed,omitempty"`
	Decimals       int                           `json:"decimals,omitempty"`
	Stats          *bondingcurvetypes.TokenStats `json:"stats,omitempty"`
}

// TokenDetails represents detailed token information
type TokenDetails struct {
	Denom       string                        `json:"denom"`
	Name        string                        `json:"name,omitempty"`
	Symbol      string                        `json:"symbol,omitempty"`
	Image       string                        `json:"image,omitempty"`
	Description string                        `json:"description,omitempty"`
	Creator     string                        `json:"creator,omitempty"`
	Decimals    int                           `json:"decimals,omitempty"`
	Stats       *bondingcurvetypes.TokenStats `json:"stats,omitempty"`
}

