package balances

import (
	"context"
	"sync"
	"time"

	"github.com/osmosis-labs/osmosis/v30/server/auth"
	"github.com/osmosis-labs/osmosis/v30/server/logger"
)

// SyncService handles periodic synchronization of balances
type SyncService struct {
	balancesService *Service
	authRepo        *auth.Repository
	log             *logger.Logger
	mu              sync.RWMutex
	activeUsers     map[int64]time.Time // user_id -> last_login_at
	activeUsersMu   sync.RWMutex
}

// NewSyncService creates a new sync service
func NewSyncService(balancesService *Service, authRepo *auth.Repository, log *logger.Logger) *SyncService {
	return &SyncService{
		balancesService: balancesService,
		authRepo:        authRepo,
		log:             log,
		activeUsers:     make(map[int64]time.Time),
	}
}

// Start starts the periodic synchronization
func (s *SyncService) Start(ctx context.Context) {
	// Load active users initially
	s.loadActiveUsers()

	// Start active users sync (every 5 minutes)
	go s.syncActiveUsers(ctx, 5*time.Minute)

	// Start full sync (every 1 hour)
	go s.syncAllUsers(ctx, 1*time.Hour)

	// Update active users list periodically (every 10 minutes)
	go s.updateActiveUsersList(ctx, 10*time.Minute)

	s.log.Info("Balance sync service started")
}

// loadActiveUsers loads active users (logged in within last 7 days)
func (s *SyncService) loadActiveUsers() {
	// Get users who logged in within last 7 days
	users, err := s.authRepo.GetActiveUsers(7 * 24 * time.Hour)
	if err != nil {
		s.log.WithError(err).Warn("Failed to load active users")
		return
	}

	s.activeUsersMu.Lock()
	defer s.activeUsersMu.Unlock()

	s.activeUsers = make(map[int64]time.Time)
	for _, user := range users {
		if user.LastLoginAt != nil {
			s.activeUsers[user.ID] = *user.LastLoginAt
		}
	}

	s.log.WithField("count", len(s.activeUsers)).Info("Loaded active users")
}

// updateActiveUsersList updates the list of active users periodically
func (s *SyncService) updateActiveUsersList(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.loadActiveUsers()
		}
	}
}

// syncActiveUsers synchronizes balances for active users
func (s *SyncService) syncActiveUsers(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.syncUsersBatch(ctx, true) // Only active users
		}
	}
}

// syncAllUsers synchronizes balances for all users
func (s *SyncService) syncAllUsers(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.syncUsersBatch(ctx, false) // All users
		}
	}
}

// syncUsersBatch synchronizes balances for a batch of users
func (s *SyncService) syncUsersBatch(ctx context.Context, activeOnly bool) {
	var users []*auth.User
	var err error

	if activeOnly {
		// Get active users
		s.activeUsersMu.RLock()
		userIDs := make([]int64, 0, len(s.activeUsers))
		for userID := range s.activeUsers {
			userIDs = append(userIDs, userID)
		}
		s.activeUsersMu.RUnlock()

		// Get full user info
		users = make([]*auth.User, 0, len(userIDs))
		for _, userID := range userIDs {
			user, err := s.authRepo.GetUserByID(userID)
			if err == nil && user != nil {
				users = append(users, user)
			}
		}
	} else {
		// Get all users with wallets
		users, err = s.authRepo.GetAllUsersWithWallets()
		if err != nil {
			s.log.WithError(err).Error("Failed to get all users with wallets")
			return
		}
	}

	if len(users) == 0 {
		return
	}

	s.log.WithField("count", len(users)).
		WithField("active_only", activeOnly).
		Info("Starting balance sync batch")

	// Process in batches of 100
	batchSize := 100
	successCount := 0
	errorCount := 0

	for i := 0; i < len(users); i += batchSize {
		end := i + batchSize
		if end > len(users) {
			end = len(users)
		}

		batch := users[i:end]
		for _, user := range batch {
			// Get wallet
			wallet, err := s.authRepo.GetWalletByUserID(user.ID)
			if err != nil || wallet == nil {
				errorCount++
				continue
			}

			// Sync balances
			if err := s.balancesService.SyncUserBalances(ctx, user.ID, wallet.Address); err != nil {
				errorCount++
				s.log.WithError(err).
					WithField("user_id", user.ID).
					WithField("address", wallet.Address).
					Warn("Failed to sync user balances")
			} else {
				successCount++
			}

			// Check context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}

	s.log.WithField("success", successCount).
		WithField("errors", errorCount).
		WithField("total", len(users)).
		Info("Balance sync batch completed")
}

