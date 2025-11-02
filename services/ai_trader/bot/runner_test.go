package bot

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/config"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/llm"
	md "github.com/osmosis-labs/osmosis/v30/services/ai_trader/marketdata"
	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/risk"
)

type noopFetcher struct{}

func (noopFetcher) GetSpot(ctx context.Context, symbol string) (md.Price, error) {
	return md.Price{}, nil
}
func (noopFetcher) GetOHLCV(ctx context.Context, symbol string, tf md.Timeframe, limit int) ([]md.Candle, error) {
	return nil, nil
}
func (noopFetcher) Name() string { return "noop" }

type fakeProvider struct{}

func (fakeProvider) GenerateDecision(ctx context.Context, in llm.PromptInput) (llm.DecisionOut, error) {
	return llm.DecisionOut{Action: "hold", Confidence: 1}, nil
}

func (fakeProvider) Name() string { return "fake" }

func newTempRepo(t *testing.T) *md.SQLiteRepository {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "runner.db")
	repo, err := md.NewSQLiteRepository(path)
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	t.Cleanup(func() {
		repo.Close()
		os.Remove(path)
	})
	return repo
}

func TestApplyPerspectiveFromRepo(t *testing.T) {
	repo := newTempRepo(t)
	market := md.NewService(noopFetcher{}).WithRepository(repo)
	provider := fakeProvider{}
	decider := risk.NewAIDecider(market, provider)
	// defaults should differ from new config to detect change
	origPre := decider.PreTF
	origTarget := decider.TargetTF
	userID, err := repo.CreateUser("bot@example.com", "bot")
	require.NoError(t, err)
	botID, err := repo.CreateBot(userID, "my-bot", "{}")
	require.NoError(t, err)
	cfg := md.PerspectiveConfig{
		PreTF:       md.TF5m,
		PreLimit:    12,
		TargetTF:    md.TF1m,
		TargetLimit: 30,
		PostTF:      md.TF1m,
		PostLimit:   20,
	}
	payload, err := cfg.Marshal()
	require.NoError(t, err)
	require.NoError(t, repo.UpdateBotConfig(botID, payload))
	c := config.DefaultConfig()
	c.Bot.Name = "my-bot"
	require.NoError(t, applyPerspectiveFromRepo(market, decider, c))
	require.Equal(t, cfg.PreTF, decider.PreTF)
	require.Equal(t, cfg.TargetTF, decider.TargetTF)
	require.Equal(t, cfg.PostTF, decider.PostTF)
	require.NotEqual(t, origPre, decider.PreTF)
	require.NotEqual(t, origTarget, decider.TargetTF)
}

func TestApplyPerspectiveFromRepo_NoMatchKeepsDefaults(t *testing.T) {
	repo := newTempRepo(t)
	market := md.NewService(noopFetcher{}).WithRepository(repo)
	decider := risk.NewAIDecider(market, fakeProvider{})
	pre := decider.PreTF
	if err := applyPerspectiveFromRepo(market, decider, config.DefaultConfig()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	require.Equal(t, pre, decider.PreTF)
}
