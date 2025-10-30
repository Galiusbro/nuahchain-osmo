package marketdata

import "time"

type Timeframe string

const (
	TF1m Timeframe = "1m"
	TF5m Timeframe = "5m"
	TF1h Timeframe = "1h"
	TF1d Timeframe = "1d"
)

type Price struct {
	Symbol    string    `json:"symbol"`
	Value     string    `json:"value"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
}

type Candle struct {
	T time.Time `json:"t"`
	O string    `json:"o"`
	H string    `json:"h"`
	L string    `json:"l"`
	C string    `json:"c"`
	V string    `json:"v"`
}

type IndicatorRequest struct {
	Symbol    string
	Timeframe Timeframe
	Window    int
}
