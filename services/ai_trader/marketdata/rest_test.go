package marketdata

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

type stubRepo struct {
	users   map[string]int64
	apiKeys map[int64]string
}

func newStubRepo() *stubRepo {
	return &stubRepo{
		users:   make(map[string]int64),
		apiKeys: make(map[int64]string),
	}
}

func (s *stubRepo) SaveLatest(Price) error                          { return nil }
func (s *stubRepo) AppendCandles(string, Timeframe, []Candle) error { return nil }
func (s *stubRepo) GetCandles(string, Timeframe, time.Time, time.Time, int) ([]Candle, error) {
	return nil, nil
}
func (s *stubRepo) SaveDecisionRecord(string, string, string, string, string, float64, string, string, string) error {
	return nil
}
func (s *stubRepo) CreateUser(email, name string) (int64, error) {
	if _, ok := s.users[email]; ok {
		return 0, errors.New("duplicate")
	}
	id := int64(len(s.users) + 1)
	s.users[email] = id
	return id, nil
}
func (s *stubRepo) GetUserByEmail(string) (*UserRecord, error) { return nil, sql.ErrNoRows }
func (s *stubRepo) CreateAPIKey(userID int64, _ string, _ *int64) (string, error) {
	key := "stubkey12345678"
	s.apiKeys[userID] = key
	return key, nil
}
func (s *stubRepo) GetAPIKeyByPrefix(int64, string) (*APIKeyRecord, error) { return nil, sql.ErrNoRows }
func (s *stubRepo) VerifyAPIKey(presented string) (*APIKeyRecord, error) {
	for uid, stored := range s.apiKeys {
		if stored == presented {
			return &APIKeyRecord{ID: 1, UserID: uid, Scope: ScopeUserAdmin, Prefix: stored[:8]}, nil
		}
	}
	return nil, errors.New("invalid")
}
func (s *stubRepo) RotateAPIKey(int64, string) (string, error) {
	return "", errors.New("not implemented")
}
func (s *stubRepo) RevokeAPIKey(int64, string) error         { return nil }
func (s *stubRepo) StoreWallet(int64, string, []byte) error  { return nil }
func (s *stubRepo) LoadWallet(int64) (string, []byte, error) { return "", nil, errors.New("not found") }
func (s *stubRepo) CreateBot(int64, string, string) (int64, error) {
	return 1, nil
}
func (s *stubRepo) GetBot(int64) (*BotRecord, error)        { return nil, sql.ErrNoRows }
func (s *stubRepo) GetBotByName(string) (*BotRecord, error) { return nil, sql.ErrNoRows }
func (s *stubRepo) ListBots(int64) ([]BotRecord, error)     { return nil, nil }
func (s *stubRepo) UpdateBotConfig(int64, string) error     { return nil }
func (s *stubRepo) CreateGrant(int64, string, string, *time.Time) (int64, error) {
	return 1, nil
}
func (s *stubRepo) LogAudit(*int64, *int64, *int64, string, string) error { return nil }
func (s *stubRepo) ListAudits(int64, int) ([]AuditRecord, error)          { return nil, nil }

func TestHandleRegisterUserWithRepo(t *testing.T) {
	t.Setenv("AI_TRADER_MASTER_KEY", "test-master-key")

	repo := newStubRepo()
	svc := NewService(NewYahooFetcher()).WithRepository(repo)
	rest := NewREST(svc)

	body, _ := json.Marshal(map[string]string{"email": "tester@example.com", "name": "Tester"})
	req := httptest.NewRequest("POST", "/users/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	rest.handleRegisterUser(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200 got %d body=%s", rec.Code, rec.Body.String())
	}
}
