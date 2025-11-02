package marketdata

import (
	"context"
	"time"
)

// Service coordinates fetchers, store, and scheduling for market data.
type Service struct {
	fetcher Fetcher
	store   *InMemoryStore
	repo    Repository
}

func NewService(fetcher Fetcher) *Service {
	return &Service{fetcher: fetcher, store: NewInMemoryStore()}
}

// WithRepository sets a persistent repository for the service.
func (s *Service) WithRepository(r Repository) *Service { s.repo = r; return s }

// Repository returns the backing repository if configured.
func (s *Service) Repository() Repository { return s.repo }

func (s *Service) Latest(ctx context.Context, symbol string) (Price, error) {
	if p, ok := s.store.GetLatest(symbol); ok && time.Since(p.Timestamp) < 30*time.Second {
		return p, nil
	}
	p, err := s.fetcher.GetSpot(ctx, symbol)
	if err != nil {
		return Price{}, err
	}
	s.store.UpsertLatest(p)
	if s.repo != nil {
		_ = s.repo.SaveLatest(p)
	}
	return p, nil
}

func (s *Service) OHLCV(ctx context.Context, symbol string, tf Timeframe, limit int) ([]Candle, error) {
	// 1) Prefer in-memory cache
	cached := s.store.GetCandles(symbol, tf, limit)
	if len(cached) >= limit {
		return cached, nil
	}
	// 2) Prefer repository to avoid Yahoo calls when we already have enough
	if s.repo != nil {
		bars, err := s.repo.GetCandles(symbol, tf, time.Time{}, time.Time{}, limit)
		if err == nil && len(bars) >= limit {
			// refresh cache and return
			s.store.AppendCandles(symbol, tf, bars)
			return s.store.GetCandles(symbol, tf, limit), nil
		}
	}
	bars, err := s.fetcher.GetOHLCV(ctx, symbol, tf, limit)
	if err != nil {
		return nil, err
	}
	s.store.AppendCandles(symbol, tf, bars)
	if s.repo != nil {
		_ = s.repo.AppendCandles(symbol, tf, bars)
	}
	return s.store.GetCandles(symbol, tf, limit), nil
}

// SaveDecisionRecord proxies to the repository if configured.
func (s *Service) SaveDecisionRecord(symbol, action, amount, paymentDenom, market string, confidence float64, rationale, promptJSON, rawResponse string) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.SaveDecisionRecord(symbol, action, amount, paymentDenom, market, confidence, rationale, promptJSON, rawResponse)
}
