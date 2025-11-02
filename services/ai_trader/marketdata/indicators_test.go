package marketdata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestIndicators_Computation(t *testing.T) {
	svc := NewService(NewYahooFetcher())
	// seed some candles via fetcher stub (returns synthetic bars)
	if _, err := svc.OHLCV(context.Background(), "AAPL", TF1m, 10); err != nil {
		t.Fatalf("seed OHLCV err: %v", err)
	}
	out, err := svc.Indicators(context.Background(), "AAPL", TF1m, []int{3}, []int{3})
	if err != nil {
		t.Fatalf("Indicators err: %v", err)
	}
	if out.MA[3] == "" || out.EMA[3] == "" {
		t.Fatalf("expected MA/EMA, got: %+v", out)
	}
}

func TestREST_IndicatorsEndpoint(t *testing.T) {
	svc := NewService(NewYahooFetcher())
	// ensure some data
	if _, err := svc.OHLCV(context.Background(), "AAPL", TF1m, 10); err != nil {
		t.Fatalf("seed OHLCV err: %v", err)
	}
	rest := NewREST(svc)
	mux := http.NewServeMux()
	rest.Register(mux)

	req := httptest.NewRequest("GET", "/indicators?"+url.Values{
		"symbol": {"AAPL"},
		"tf":     {"1m"},
		"ma":     {"3"},
		"ema":    {"3"},
	}.Encode(), nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("unexpected status: %d body=%s", w.Code, w.Body.String())
	}
	if len(w.Body.String()) == 0 {
		t.Fatalf("empty body")
	}
}
