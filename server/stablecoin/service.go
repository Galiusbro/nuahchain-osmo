package stablecoin

import (
	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/blockchain"
)

// Service handles stablecoin operations
type Service struct {
	authService   *auth.Service
	blockchainCli *blockchain.Client
}

// NewService creates a new stablecoin service
func NewService(authService *auth.Service, blockchainCli *blockchain.Client) *Service {
	return &Service{
		authService:   authService,
		blockchainCli: blockchainCli,
	}
}

var globalService *Service
var globalAuthService *auth.Service

// SetService sets the global service instance
func SetService(s *Service) {
	globalService = s
}

// SetAuthService sets the global auth service instance
func SetAuthService(s *auth.Service) {
	globalAuthService = s
}
