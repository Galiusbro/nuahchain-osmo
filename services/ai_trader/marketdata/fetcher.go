package marketdata

import "context"

// Fetcher fetches latest spot and OHLCV from an external source.
type Fetcher interface {
	GetSpot(ctx context.Context, symbol string) (Price, error)
	GetOHLCV(ctx context.Context, symbol string, tf Timeframe, limit int) ([]Candle, error)
	Name() string
}
