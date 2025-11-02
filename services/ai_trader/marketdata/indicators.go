package marketdata

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"
)

type IndicatorResult struct {
	MA         map[int]string `json:"ma,omitempty"`
	EMA        map[int]string `json:"ema,omitempty"`
	RSI        string         `json:"rsi,omitempty"`
	MACD       string         `json:"macd,omitempty"`
	MACDSignal string         `json:"macd_signal,omitempty"`
	MACDHist   string         `json:"macd_hist,omitempty"`
	ATR        string         `json:"atr,omitempty"`
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
	// RSI(14), MACD(12,26,9), ATR(14)
	out.RSI = rsi(candles, 14)
	macdVal, macdSig, macdHist := macd(closes, 12, 26, 9)
	out.MACD, out.MACDSignal, out.MACDHist = macdVal, macdSig, macdHist
	out.ATR = atr(candles, 14)
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

// rsi computes RSI using Wilder's smoothing
func rsi(candles []Candle, period int) string {
	if period <= 0 || len(candles) < period+1 {
		return ""
	}
	gains, losses := 0.0, 0.0
	for i := 1; i <= period; i++ {
		change := atof(candles[i].C) - atof(candles[i-1].C)
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	for i := period + 1; i < len(candles); i++ {
		change := atof(candles[i].C) - atof(candles[i-1].C)
		gain, loss := 0.0, 0.0
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
	}
	if avgLoss == 0 {
		return "100.000000"
	}
	rs := avgGain / avgLoss
	r := 100 - (100 / (1 + rs))
	return ftoa(r)
}

// macd returns MACD line, signal, histogram
func macd(closes []string, fast, slow, signal int) (string, string, string) {
	if len(closes) < slow+signal {
		return "", "", ""
	}
	// helper ema float series
	toFloat := func(n int) []float64 {
		arr := make([]float64, len(closes))
		for i := range closes {
			arr[i] = atof(closes[i])
		}
		return arr
	}
	series := toFloat(len(closes))
	emaN := func(src []float64, n int) []float64 {
		k := 2.0 / (float64(n) + 1)
		out := make([]float64, len(src))
		sum := 0.0
		for i := 0; i < n; i++ {
			sum += src[i]
		}
		out[n-1] = sum / float64(n)
		for i := n; i < len(src); i++ {
			out[i] = src[i]*k + out[i-1]*(1-k)
		}
		return out
	}
	emaFast := emaN(series, fast)
	emaSlow := emaN(series, slow)
	macdLine := make([]float64, len(series))
	for i := 0; i < len(series); i++ {
		macdLine[i] = emaFast[i] - emaSlow[i]
	}
	sig := emaN(macdLine, signal)
	hist := macdLine[len(macdLine)-1] - sig[len(sig)-1]
	return ftoa(macdLine[len(macdLine)-1]), ftoa(sig[len(sig)-1]), ftoa(hist)
}

// atr computes Average True Range
func atr(candles []Candle, period int) string {
	if period <= 0 || len(candles) < period+1 {
		return ""
	}
	trs := make([]float64, 0, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		h := atof(candles[i].H)
		l := atof(candles[i].L)
		pc := atof(candles[i-1].C)
		tr := math.Max(h-l, math.Max(math.Abs(h-pc), math.Abs(l-pc)))
		trs = append(trs, tr)
	}
	// Wilder's smoothing
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += trs[i]
	}
	atr := sum / float64(period)
	for i := period; i < len(trs); i++ {
		atr = (atr*float64(period-1) + trs[i]) / float64(period)
	}
	return ftoa(atr)
}
