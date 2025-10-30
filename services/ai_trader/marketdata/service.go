package marketdata

import (
	"context"
	"time"
)

// Service coordinates fetchers, store, and scheduling for market data.
type Service struct {
	fetcher Fetcher
	store   *InMemoryStore
}

func NewService(fetcher Fetcher) *Service {
	return &Service{fetcher: fetcher, store: NewInMemoryStore()}
}

func (s *Service) Latest(ctx context.Context, symbol string) (Price, error) {
	if p, ok := s.store.GetLatest(symbol); ok && time.Since(p.Timestamp) < 30*time.Second {
		return p, nil
	}
	p, err := s.fetcher.GetSpot(ctx, symbol)
	if err != nil {
		return Price{}, err
	}
	s.store.UpsertLatest(p)
	return p, nil
}

func (s *Service) OHLCV(ctx context.Context, symbol string, tf Timeframe, limit int) ([]Candle, error) {
	cached := s.store.GetCandles(symbol, tf, limit)
	if len(cached) >= limit {
		return cached, nil
	}
	bars, err := s.fetcher.GetOHLCV(ctx, symbol, tf, limit)
	if err != nil {
		return nil, err
	}
	s.store.AppendCandles(symbol, tf, bars)
	return s.store.GetCandles(symbol, tf, limit), nil
}
