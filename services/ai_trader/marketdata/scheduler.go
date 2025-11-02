package marketdata

import (
	"context"
	"time"
)

// Scheduler periodically refreshes latest price and OHLCV for a symbol set.
type Scheduler struct {
	Service  *Service
	Symbols  []string
	TFs      []Timeframe
	Interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	initial  map[string]bool // per-symbol initial backfill flag
}

func NewScheduler(svc *Service, symbols []string, tfs []Timeframe, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	init := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		init[s] = false
	}
	return &Scheduler{Service: svc, Symbols: symbols, TFs: tfs, Interval: interval, ctx: ctx, cancel: cancel, initial: init}
}

func (s *Scheduler) Start() {
	go func() {
		ticker := time.NewTicker(s.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.refreshOnce()
			}
		}
	}()
}

func (s *Scheduler) Stop() { s.cancel() }

func (s *Scheduler) refreshOnce() {
	ctx, cancel := context.WithTimeout(s.ctx, s.Interval)
	defer cancel()
	for _, sym := range s.Symbols {
		// Always update latest spot
		_, _ = s.Service.Latest(ctx, sym)
		// Backfill once with large windows, then small incremental
		initial := !s.initial[sym]
		for _, tf := range s.TFs {
			limit := 100
			switch tf {
			case TF1m:
				if initial {
					limit = 300
				} else {
					limit = 120
				}
			case TF5m:
				if initial {
					limit = 1000
				} else {
					limit = 120
				}
			case TF1h:
				if initial {
					limit = 720
				} else {
					limit = 48
				}
			case TF1d:
				if initial {
					limit = 365
				} else {
					limit = 7
				}
			}
			_, _ = s.Service.OHLCV(ctx, sym, tf, limit)
			// small delay to respect provider rate limits across multiple intervals
			time.Sleep(2 * time.Second)
		}
		if initial {
			s.initial[sym] = true
		}
	}
}
