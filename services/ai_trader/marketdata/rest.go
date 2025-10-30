package marketdata

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// REST exposes minimal HTTP endpoints for market data.
type REST struct {
	svc *Service
}

func NewREST(svc *Service) *REST { return &REST{svc: svc} }

func (r *REST) Register(mux *http.ServeMux) {
	mux.HandleFunc("/spot", r.handleSpot)
	mux.HandleFunc("/ohlcv", r.handleOHLCV)
	mux.HandleFunc("/indicators", r.handleIndicators)
}

func (r *REST) handleSpot(w http.ResponseWriter, req *http.Request) {
	sym := req.URL.Query().Get("symbol")
	if sym == "" {
		http.Error(w, "symbol required", http.StatusBadRequest)
		return
	}
	p, err := r.svc.Latest(req.Context(), sym)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"symbol": p.Symbol, "price": p.Value, "source": p.Source, "ts": p.Timestamp.Unix()})
}

func (r *REST) handleOHLCV(w http.ResponseWriter, req *http.Request) {
	sym := req.URL.Query().Get("symbol")
	tf := Timeframe(req.URL.Query().Get("tf"))
	if tf == "" {
		tf = TF1m
	}
	if sym == "" {
		http.Error(w, "symbol required", http.StatusBadRequest)
		return
	}
	bars, err := r.svc.OHLCV(req.Context(), sym, tf, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	_ = json.NewEncoder(w).Encode(bars)
}

func (r *REST) handleIndicators(w http.ResponseWriter, req *http.Request) {
	sym := req.URL.Query().Get("symbol")
	tf := Timeframe(req.URL.Query().Get("tf"))
	if tf == "" {
		tf = TF1m
	}
	var maWins, emaWins []int
	if ma := strings.TrimSpace(req.URL.Query().Get("ma")); ma != "" {
		for _, p := range strings.Split(ma, ",") {
			if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				maWins = append(maWins, n)
			}
		}
	}
	if ema := strings.TrimSpace(req.URL.Query().Get("ema")); ema != "" {
		for _, p := range strings.Split(ema, ",") {
			if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				emaWins = append(emaWins, n)
			}
		}
	}
	out, err := r.svc.Indicators(req.Context(), sym, tf, maWins, emaWins)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}
