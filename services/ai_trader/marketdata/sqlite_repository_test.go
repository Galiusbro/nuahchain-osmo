package marketdata

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func setTestMasterKey(t *testing.T) {
	// helper to ensure deterministic key derivation for tests
	t.Helper()
	prev := os.Getenv("AI_TRADER_MASTER_KEY")
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 32))
	if err := os.Setenv("AI_TRADER_MASTER_KEY", key); err != nil {
		t.Fatalf("set env: %v", err)
	}
	t.Cleanup(func() {
		if prev == "" {
			_ = os.Unsetenv("AI_TRADER_MASTER_KEY")
			return
		}
		_ = os.Setenv("AI_TRADER_MASTER_KEY", prev)
	})
}

func newTestRepo(t *testing.T) *SQLiteRepository {
	t.Helper()
	setTestMasterKey(t)
	dir := t.TempDir()
	dsn := filepath.Join(dir, "repo.db")
	repo, err := NewSQLiteRepository(dsn)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	return repo
}

func TestSQLiteRepository_APIKeyLifecycle(t *testing.T) {
	repo := newTestRepo(t)
	userID, err := repo.CreateUser("alice@example.com", "Alice")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := repo.CreateAPIKey(userID, "bad_scope", nil); err == nil {
		t.Fatal("expected error for unknown scope")
	}
	if _, err := repo.CreateAPIKey(userID, ScopeBotTrade, nil); err == nil {
		t.Fatal("expected error when bot trade scope missing bot_id")
	}
	key, err := repo.CreateAPIKey(userID, ScopeUserAdmin, nil)
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}
	if key == "" {
		t.Fatal("expected key value")
	}
	rec, err := repo.VerifyAPIKey(key)
	if err != nil {
		t.Fatalf("verify key: %v", err)
	}
	if rec.UserID != userID {
		t.Fatalf("user mismatch: %d", rec.UserID)
	}
	if rec.Scope != ScopeUserAdmin {
		t.Fatalf("scope mismatch: %s", rec.Scope)
	}
	if rec.BotID != nil {
		t.Fatalf("expected nil bot for user key")
	}
	if fetched, err := repo.GetAPIKeyByPrefix(userID, rec.Prefix); err != nil {
		t.Fatalf("get by prefix: %v", err)
	} else if fetched.ID != rec.ID {
		t.Fatalf("get by prefix mismatch: %d", fetched.ID)
	}
	oldPrefix := rec.Prefix
	rotated, err := repo.RotateAPIKey(userID, oldPrefix)
	if err != nil {
		t.Fatalf("rotate key: %v", err)
	}
	if rotated == key {
		t.Fatal("expected different key after rotation")
	}
	if _, err := repo.VerifyAPIKey(key); err == nil {
		t.Fatal("expected old key to be revoked")
	}
	rec2, err := repo.VerifyAPIKey(rotated)
	if err != nil {
		t.Fatalf("verify rotated: %v", err)
	}
	if rec2.Scope != ScopeUserAdmin {
		t.Fatalf("rotated key scope mismatch: %s", rec2.Scope)
	}
	if _, err := repo.GetAPIKeyByPrefix(userID, oldPrefix); err == nil {
		t.Fatal("expected old prefix lookup to fail")
	}
	if fetched, err := repo.GetAPIKeyByPrefix(userID, rec2.Prefix); err != nil {
		t.Fatalf("get rotated by prefix: %v", err)
	} else if fetched.ID != rec2.ID {
		t.Fatalf("rotated prefix mismatch: %d", fetched.ID)
	}
	botID, err := repo.CreateBot(userID, "bot-one", "{}")
	if err != nil {
		t.Fatalf("create bot: %v", err)
	}
	if _, err := repo.CreateAPIKey(userID, ScopeUserRead, &botID); err == nil {
		t.Fatal("expected error when user scope provided bot_id")
	}
	otherUserID, err := repo.CreateUser("bob@example.com", "Bob")
	if err != nil {
		t.Fatalf("create other user: %v", err)
	}
	if _, err := repo.CreateAPIKey(otherUserID, ScopeBotTrade, &botID); err == nil {
		t.Fatal("expected error for bot owner mismatch")
	}
	botKey, err := repo.CreateAPIKey(userID, ScopeBotTrade, &botID)
	if err != nil {
		t.Fatalf("create bot api key: %v", err)
	}
	recBot, err := repo.VerifyAPIKey(botKey)
	if err != nil {
		t.Fatalf("verify bot key: %v", err)
	}
	if recBot.BotID == nil || *recBot.BotID != botID {
		t.Fatalf("expected bot id %d, got %+v", botID, recBot.BotID)
	}
	if recBot.Scope != ScopeBotTrade {
		t.Fatalf("expected bot scope, got %s", recBot.Scope)
	}
	botRotated, err := repo.RotateAPIKey(userID, recBot.Prefix)
	if err != nil {
		t.Fatalf("rotate bot key: %v", err)
	}
	if _, err := repo.VerifyAPIKey(botKey); err == nil {
		t.Fatal("expected old bot key revoked")
	}
	rotRec, err := repo.VerifyAPIKey(botRotated)
	if err != nil {
		t.Fatalf("verify rotated bot key: %v", err)
	}
	if rotRec.BotID == nil || *rotRec.BotID != botID {
		t.Fatalf("rotated bot key lost bot binding")
	}
	if rotRec.Scope != ScopeBotTrade {
		t.Fatalf("rotated bot key scope mismatch: %s", rotRec.Scope)
	}
	if err := repo.RevokeAPIKey(userID, rotRec.Prefix); err != nil {
		t.Fatalf("revoke bot key: %v", err)
	}
	if err := repo.RevokeAPIKey(userID, rotRec.Prefix); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows on double revoke, got %v", err)
	}
}

func TestSQLiteRepository_WalletEncryption(t *testing.T) {
	repo := newTestRepo(t)
	userID, err := repo.CreateUser("wallet@example.com", "Wallet User")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	plaintxt := []byte("secret-private-key")
	if err := repo.StoreWallet(userID, "nuah1address", plaintxt); err != nil {
		t.Fatalf("store wallet: %v", err)
	}
	addr, recovered, err := repo.LoadWallet(userID)
	if err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if addr != "nuah1address" {
		t.Fatalf("address mismatch: %s", addr)
	}
	if !bytes.Equal(recovered, plaintxt) {
		t.Fatalf("wallet decrypt mismatch")
	}
	row := repo.db.QueryRow(`SELECT priv_enc FROM wallets WHERE user_id=?`, userID)
	var stored []byte
	if err := row.Scan(&stored); err != nil {
		t.Fatalf("scan priv: %v", err)
	}
	if bytes.Equal(stored, plaintxt) {
		t.Fatalf("expected encrypted data, stored matches plaintext")
	}
}

func TestSQLiteRepository_StoreWalletReassignsAddress(t *testing.T) {
	repo := newTestRepo(t)
	firstUser, err := repo.CreateUser("first@example.com", "First")
	if err != nil {
		t.Fatalf("create first user: %v", err)
	}
	secondUser, err := repo.CreateUser("second@example.com", "Second")
	if err != nil {
		t.Fatalf("create second user: %v", err)
	}
	address := "nuah1sharedaddr"

	if err := repo.StoreWallet(firstUser, address, []byte("first-key")); err != nil {
		t.Fatalf("store first wallet: %v", err)
	}
	// Second user overwrites same address
	if err := repo.StoreWallet(secondUser, address, []byte("second-key")); err != nil {
		t.Fatalf("store second wallet: %v", err)
	}

	// First user should no longer see the wallet
	if _, _, err := repo.LoadWallet(firstUser); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for first user, got %v", err)
	}

	addr, key, err := repo.LoadWallet(secondUser)
	if err != nil {
		t.Fatalf("load wallet second user: %v", err)
	}
	if addr != address {
		t.Fatalf("expected address %s, got %s", address, addr)
	}
	if string(key) != "second-key" {
		t.Fatalf("expected second key, got %q", string(key))
	}
}

func TestSQLiteRepository_AuditLogging(t *testing.T) {
	repo := newTestRepo(t)
	userID, err := repo.CreateUser("audit@example.com", "Auditor")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	key, err := repo.CreateAPIKey(userID, ScopeUserAdmin, nil)
	if err != nil {
		t.Fatalf("create key: %v", err)
	}
	rec, err := repo.VerifyAPIKey(key)
	if err != nil {
		t.Fatalf("verify key: %v", err)
	}
	uid := rec.UserID
	if err := repo.LogAudit(&uid, rec.BotID, &rec.ID, "api.login", `{"ok":true}`); err != nil {
		t.Fatalf("log audit: %v", err)
	}
	for i := 0; i < 3; i++ {
		action := fmt.Sprintf("event-%d", i)
		if err := repo.LogAudit(&uid, nil, &rec.ID, action, ""); err != nil {
			t.Fatalf("log audit loop: %v", err)
		}
	}
	entries, err := repo.ListAudits(userID, 2)
	if err != nil {
		t.Fatalf("list audits: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Action != "event-2" {
		t.Fatalf("expected latest action event-2, got %s", entries[0].Action)
	}
	if entries[0].APIKeyID == nil || *entries[0].APIKeyID != rec.ID {
		t.Fatalf("api key linkage mismatch")
	}
	if entries[1].MetaJSON != "{}" {
		t.Fatalf("expected default meta json, got %s", entries[1].MetaJSON)
	}
}
