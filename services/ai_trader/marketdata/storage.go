package marketdata

import (
	"sync"
	"time"
)

// InMemoryStore holds latest prices and OHLCV history in memory.
type InMemoryStore struct {
	mu      sync.RWMutex
	latest  map[string]Price                  // symbol -> price
	candles map[string]map[Timeframe][]Candle // symbol -> tf -> candles (ascending by time)
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		latest:  make(map[string]Price),
		candles: make(map[string]map[Timeframe][]Candle),
	}
}

func (s *InMemoryStore) UpsertLatest(p Price) {
	s.mu.Lock()
	s.latest[p.Symbol] = p
	s.mu.Unlock()
}

func (s *InMemoryStore) GetLatest(symbol string) (Price, bool) {
	s.mu.RLock()
	p, ok := s.latest[symbol]
	s.mu.RUnlock()
	return p, ok
}

func (s *InMemoryStore) AppendCandles(symbol string, tf Timeframe, bars []Candle) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.candles[symbol]
	if !ok {
		m = make(map[Timeframe][]Candle)
		s.candles[symbol] = m
	}
	m[tf] = append(m[tf], bars...)
}

func (s *InMemoryStore) GetCandles(symbol string, tf Timeframe, limit int) []Candle {
	s.mu.RLock()
	arr := s.candles[symbol][tf]
	s.mu.RUnlock()
	if limit <= 0 || len(arr) <= limit {
		return arr
	}
	return arr[len(arr)-limit:]
}

// PruneOld keeps only candles newer than cutoff.
func (s *InMemoryStore) PruneOld(cutoff time.Time) {
	s.mu.Lock()
	for sym, tfm := range s.candles {
		for tf, arr := range tfm {
			kept := arr[:0]
			for _, c := range arr {
				if !c.T.Before(cutoff) {
					kept = append(kept, c)
				}
			}
			tfm[tf] = kept
		}
		s.candles[sym] = tfm
	}
	s.mu.Unlock()
}
