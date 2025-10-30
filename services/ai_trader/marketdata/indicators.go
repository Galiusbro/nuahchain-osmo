package marketdata

import (
	"context"
	"errors"
	"strconv"
	"time"
)

type IndicatorResult struct {
	MA  map[int]string `json:"ma,omitempty"`
	EMA map[int]string `json:"ema,omitempty"`
}

// Indicators computes requested indicators using repo/store data.
func (s *Service) Indicators(ctx context.Context, symbol string, tf Timeframe, maWindows, emaWindows []int) (IndicatorResult, error) {
	if symbol == "" {
		return IndicatorResult{}, errors.New("symbol required")
	}
	// Pull up to the max window + safety margin
	maxN := max(sliceMax(maWindows), sliceMax(emaWindows))
	if maxN <= 0 {
		maxN = 50
	}
	var candles []Candle
	// Prefer repo if available for consistent history
	if s.repo != nil {
		from := time.Now().UTC().Add(-time.Duration(maxN+10) * time.Hour) // rough window
		got, err := s.repo.GetCandles(symbol, tf, from, time.Now().UTC(), maxN+10)
		if err == nil && len(got) > 0 {
			candles = got
		}
	}
	if len(candles) == 0 {
		// fallback to in-memory store
		candles = s.store.GetCandles(symbol, tf, maxN+10)
	}
	if len(candles) == 0 {
		return IndicatorResult{}, errors.New("no candles")
	}

	closes := make([]string, 0, len(candles))
	for _, c := range candles {
		closes = append(closes, c.C)
	}

	out := IndicatorResult{MA: map[int]string{}, EMA: map[int]string{}}
	for _, w := range maWindows {
		if w > 0 {
			out.MA[w] = sma(closes, w)
		}
	}
	for _, w := range emaWindows {
		if w > 0 {
			out.EMA[w] = ema(closes, w)
		}
	}
	return out, nil
}

func sma(closes []string, window int) string {
	if window <= 0 || len(closes) < window {
		return ""
	}
	// Use decimal via float64 for simplicity; can switch to osmomath later
	var sum float64
	for i := len(closes) - window; i < len(closes); i++ {
		sum += atof(closes[i])
	}
	return ftoa(sum / float64(window))
}

func ema(closes []string, window int) string {
	if window <= 0 || len(closes) < window {
		return ""
	}
	k := 2.0 / (float64(window) + 1.0)
	// seed with SMA of first window
	var seed float64
	for i := 0; i < window; i++ {
		seed += atof(closes[i])
	}
	ema := seed / float64(window)
	for i := window; i < len(closes); i++ {
		price := atof(closes[i])
		ema = price*k + ema*(1.0-k)
	}
	return ftoa(ema)
}

// helpers
func atof(s string) float64 { x, _ := strconv.ParseFloat(s, 64); return x }
func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 6, 64) }
func sliceMax(arr []int) int {
	m := 0
	for _, v := range arr {
		if v > m {
			m = v
		}
	}
	return m
}
