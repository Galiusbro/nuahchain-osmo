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
}

func NewScheduler(svc *Service, symbols []string, tfs []Timeframe, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{Service: svc, Symbols: symbols, TFs: tfs, Interval: interval, ctx: ctx, cancel: cancel}
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
		_, _ = s.Service.Latest(ctx, sym)
		for _, tf := range s.TFs {
			_, _ = s.Service.OHLCV(ctx, sym, tf, 50)
		}
	}
}
