package marketdata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

type stubFetcher struct{}

func (stubFetcher) GetSpot(ctx context.Context, symbol string) (Price, error) {
	return Price{
		Symbol:    symbol,
		Value:     "123.45",
		Source:    "stub",
		Timestamp: time.Now().UTC(),
	}, nil
}

func (stubFetcher) GetOHLCV(ctx context.Context, symbol string, tf Timeframe, limit int) ([]Candle, error) {
	return nil, nil
}

func (stubFetcher) Name() string { return "stub_fetcher" }

func TestREST_RegisterAndStoreWallet(t *testing.T) {
	t.Setenv("AI_TRADER_MASTER_KEY", "rest-integration-secret")

	tmpDir := t.TempDir()
	dsn := fmt.Sprintf("file:%s?cache=shared&_journal_mode=WAL", filepath.Join(tmpDir, "rest_integration.db"))

	repo, err := NewSQLiteRepository(dsn)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	svc := NewService(stubFetcher{}).WithRepository(repo)
	rest := NewREST(svc)

	mux := http.NewServeMux()
	rest.Register(mux)

	server := httptest.NewServer(mux)
	defer server.Close()

	email := fmt.Sprintf("rest-%d@example.com", time.Now().UnixNano())
	registerPayload, _ := json.Marshal(map[string]string{
		"email": email,
		"name":  "REST Integration",
	})

	resp, err := http.Post(server.URL+"/users/register", "application/json", bytes.NewReader(registerPayload))
	if err != nil {
		t.Fatalf("register request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read register body: %v", err)
	}
	var reg registerResp
	if err := json.Unmarshal(body, &reg); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if reg.APIKey == "" {
		t.Fatalf("empty api key in response")
	}

	walletPayload, _ := json.Marshal(map[string]string{
		"address":     "nuah1restintegration",
		"priv_base64": "c3R1Yl9wcml2YXRlX2tleQ==",
	})

	req, err := http.NewRequest(http.MethodPost, server.URL+"/wallets/create", bytes.NewReader(walletPayload))
	if err != nil {
		t.Fatalf("create wallet request: %v", err)
	}
	req.Header.Set("X-API-Key", reg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("wallet request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("wallet status %d", resp.StatusCode)
	}

	userRec, err := repo.GetUserByEmail(email)
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}
	if userRec.Email != email {
		t.Fatalf("email mismatch %s", userRec.Email)
	}

	addr, priv, err := repo.LoadWallet(userRec.ID)
	if err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if addr != "nuah1restintegration" {
		t.Fatalf("wallet address mismatch %s", addr)
	}
	if string(priv) != "stub_private_key" {
		t.Fatalf("wallet payload mismatch")
	}
}
