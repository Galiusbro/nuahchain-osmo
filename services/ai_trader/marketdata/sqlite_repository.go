package marketdata

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Repository abstracts persistence for prices and candles.
type Repository interface {
	SaveLatest(p Price) error
	AppendCandles(symbol string, tf Timeframe, bars []Candle) error
	GetCandles(symbol string, tf Timeframe, from, to time.Time, limit int) ([]Candle, error)
	SaveDecisionRecord(symbol, action, amount, paymentDenom, market string, confidence float64, rationale, promptJSON, rawResponse string) error
	// identity and auth
	CreateUser(email, name string) (int64, error)
	GetUserByEmail(email string) (*UserRecord, error)
	CreateAPIKey(userID int64, scope string, botID *int64) (string, error)
	GetAPIKeyByPrefix(userID int64, prefix string) (*APIKeyRecord, error)
	VerifyAPIKey(presented string) (*APIKeyRecord, error)
	RotateAPIKey(userID int64, oldPrefix string) (string, error)
	RevokeAPIKey(userID int64, prefix string) error
	StoreWallet(userID int64, address string, privPlain []byte) error
	LoadWallet(userID int64) (address string, privPlain []byte, err error)
	CreateBot(userID int64, name string, configJSON string) (int64, error)
	GetBot(botID int64) (*BotRecord, error)
	GetBotByName(name string) (*BotRecord, error)
	ListBots(userID int64) ([]BotRecord, error)
	UpdateBotConfig(botID int64, configJSON string) error
	CreateGrant(botID int64, grantType string, paramsJSON string, expiresAt *time.Time) (int64, error)
	LogAudit(userID *int64, botID *int64, apiKeyID *int64, action string, metaJSON string) error
	ListAudits(userID int64, limit int) ([]AuditRecord, error)
}

// UserRecord represents a persisted user row.
type UserRecord struct {
	ID        int64
	Email     string
	Name      string
	CreatedAt time.Time
}

// APIKeyRecord captures metadata about an API key, including scope and bot association.
type APIKeyRecord struct {
	ID         int64
	UserID     int64
	BotID      *int64
	Scope      string
	Prefix     string
	LastUsedAt *time.Time
}

// BotRecord describes a trading bot owned by a user.
type BotRecord struct {
	ID         int64
	UserID     int64
	Name       string
	ConfigJSON string
	CreatedAt  time.Time
}

// AuditRecord captures a single audit trail entry.
type AuditRecord struct {
	ID        int64
	UserID    *int64
	BotID     *int64
	APIKeyID  *int64
	Action    string
	MetaJSON  string
	CreatedAt time.Time
}

func ensureColumn(db *sql.DB, table, column, ddl string) error {
	exists, err := columnExists(db, table, column)
	if err != nil || exists {
		return err
	}
	stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", table, ddl)
	_, err = db.Exec(stmt)
	return err
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", table)
	rows, err := db.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var ignore1, ignore2, ignore3, ignore4, ignore5 interface{}
		if err := rows.Scan(&ignore1, &name, &ignore2, &ignore3, &ignore4, &ignore5); err != nil {
			return false, err
		}
		if strings.EqualFold(name, column) {
			return true, nil
		}
	}
	return false, rows.Err()
}

func nullableInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func int64PtrFromNull(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

func timePtrFromNull(n sql.NullInt64) *time.Time {
	if !n.Valid {
		return nil
	}
	t := time.Unix(n.Int64, 0).UTC()
	return &t
}

type SQLiteRepository struct {
	db      *sql.DB
	writeMu sync.Mutex
}

func NewSQLiteRepository(dsn string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// Limit concurrent connections; SQLite allows a single writer, many readers.
	// Keeping this low avoids writer starvation/timeouts under load.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA foreign_keys=ON; PRAGMA busy_timeout=10000;`); err != nil {
		return nil, err
	}
	// Ensure DB is reachable quickly to avoid hanging tests
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	r := &SQLiteRepository{db: db}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

// withWrite serializes mutating operations to avoid SQLITE_BUSY/LOCKED under load.
func (r *SQLiteRepository) withWrite(fn func() error) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()
	return fn()
}

func (r *SQLiteRepository) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS prices (
  symbol TEXT PRIMARY KEY,
  value  TEXT NOT NULL,
  source TEXT NOT NULL,
  ts_unix INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS candles (
  symbol TEXT NOT NULL,
  timeframe TEXT NOT NULL,
  ts_unix INTEGER NOT NULL,
  o TEXT NOT NULL,
  h TEXT NOT NULL,
  l TEXT NOT NULL,
  c TEXT NOT NULL,
  v TEXT NOT NULL,
  PRIMARY KEY(symbol, timeframe, ts_unix)
);
CREATE INDEX IF NOT EXISTS idx_candles_symbol_tf_ts ON candles(symbol, timeframe, ts_unix);
CREATE TABLE IF NOT EXISTS decisions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  symbol TEXT NOT NULL,
  action TEXT NOT NULL,
  amount TEXT NOT NULL,
  payment_denom TEXT NOT NULL,
  market TEXT NOT NULL,
  confidence REAL NOT NULL,
  rationale TEXT NOT NULL,
  prompt_json TEXT NOT NULL,
  raw_response TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
-- Identity and authorization
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL,
  name TEXT,
  created_at INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE TABLE IF NOT EXISTS bots (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  config_json TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_bots_user ON bots(user_id);
CREATE TABLE IF NOT EXISTS api_keys (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  bot_id INTEGER,
  prefix TEXT NOT NULL,
  key_hash TEXT NOT NULL,
  secret_enc BLOB NOT NULL,
  scope TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  revoked_at INTEGER,
  last_used_at INTEGER,
  FOREIGN KEY(user_id) REFERENCES users(id),
  FOREIGN KEY(bot_id) REFERENCES bots(id)
);
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_bot ON api_keys(bot_id);
CREATE TABLE IF NOT EXISTS wallets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  address TEXT NOT NULL,
  priv_enc BLOB NOT NULL,
  created_at INTEGER NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address);
CREATE TABLE IF NOT EXISTS grants (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  bot_id INTEGER NOT NULL,
  grant_type TEXT NOT NULL,
  params_json TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  expires_at INTEGER,
  revoked_at INTEGER,
  FOREIGN KEY(bot_id) REFERENCES bots(id)
);
CREATE INDEX IF NOT EXISTS idx_grants_bot ON grants(bot_id);
CREATE INDEX IF NOT EXISTS idx_grants_bot_expires ON grants(bot_id, expires_at);
CREATE TABLE IF NOT EXISTS audits (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER,
  bot_id INTEGER,
  api_key_id INTEGER,
  action TEXT NOT NULL,
  meta_json TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id),
  FOREIGN KEY(bot_id) REFERENCES bots(id),
  FOREIGN KEY(api_key_id) REFERENCES api_keys(id)
);
CREATE INDEX IF NOT EXISTS idx_audits_user ON audits(user_id);
CREATE INDEX IF NOT EXISTS idx_audits_api_key ON audits(api_key_id);
`

	if _, err := r.db.Exec(schema); err != nil {
		return err
	}
	if err := ensureColumn(r.db, "api_keys", "bot_id", "bot_id INTEGER"); err != nil {
		return err
	}
	if err := ensureColumn(r.db, "api_keys", "secret_enc", "secret_enc BLOB NOT NULL DEFAULT X''"); err != nil {
		return err
	}
	if err := ensureColumn(r.db, "api_keys", "scope", "scope TEXT NOT NULL DEFAULT 'user_admin'"); err != nil {
		return err
	}
	if err := ensureColumn(r.db, "api_keys", "last_used_at", "last_used_at INTEGER"); err != nil {
		return err
	}
	if err := ensureColumn(r.db, "grants", "revoked_at", "revoked_at INTEGER"); err != nil {
		return err
	}
	if err := ensureColumn(r.db, "audits", "bot_id", "bot_id INTEGER"); err != nil {
		return err
	}
	return nil
}

func (r *SQLiteRepository) SaveLatest(p Price) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()
	_, err := r.db.Exec(`INSERT INTO prices(symbol,value,source,ts_unix) VALUES(?,?,?,?)
ON CONFLICT(symbol) DO UPDATE SET value=excluded.value, source=excluded.source, ts_unix=excluded.ts_unix`, p.Symbol, p.Value, p.Source, p.Timestamp.Unix())
	return err
}

func (r *SQLiteRepository) AppendCandles(symbol string, tf Timeframe, bars []Candle) error {
	if len(bars) == 0 {
		return nil
	}
	r.writeMu.Lock()
	defer r.writeMu.Unlock()
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO candles(symbol,timeframe,ts_unix,o,h,l,c,v) VALUES(?,?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, c := range bars {
		if _, err := stmt.Exec(symbol, string(tf), c.T.Unix(), c.O, c.H, c.L, c.C, c.V); err != nil {
			stmt.Close()
			tx.Rollback()
			return err
		}
	}
	stmt.Close()
	return tx.Commit()
}

func (r *SQLiteRepository) GetCandles(symbol string, tf Timeframe, from, to time.Time, limit int) ([]Candle, error) {
	if to.IsZero() {
		to = time.Now().UTC()
	}
	if from.After(to) {
		return nil, errors.New("from > to")
	}
	rows, err := r.db.Query(`SELECT ts_unix,o,h,l,c,v FROM candles WHERE symbol=? AND timeframe=? AND ts_unix BETWEEN ? AND ? ORDER BY ts_unix ASC`, symbol, string(tf), from.Unix(), to.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Candle, 0, 256)
	for rows.Next() {
		var ts int64
		var o, h, l, c, v string
		if err := rows.Scan(&ts, &o, &h, &l, &c, &v); err != nil {
			return nil, err
		}
		out = append(out, Candle{T: time.Unix(ts, 0).UTC(), O: o, H: h, L: l, C: c, V: v})
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, rows.Err()
}

func (r *SQLiteRepository) Close() error { return r.db.Close() }

// Optional: convenience to set context deadline on DB ops (example usage)
func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if _, has := ctx.Deadline(); has {
		return ctx, func() {}
	}
	if d <= 0 {
		d = 5 * time.Second
	}
	return context.WithTimeout(ctx, d)
}

func (r *SQLiteRepository) SaveDecisionRecord(symbol, action, amount, paymentDenom, market string, confidence float64, rationale, promptJSON, rawResponse string) error {
	return r.withWrite(func() error {
		_, err := r.db.Exec(`INSERT INTO decisions(symbol,action,amount,payment_denom,market,confidence,rationale,prompt_json,raw_response,created_at)
VALUES(?,?,?,?,?,?,?,?,?,?)`, symbol, action, amount, paymentDenom, market, confidence, rationale, promptJSON, rawResponse, time.Now().UTC().Unix())
		return err
	})
}

// Identity and API keys
func (r *SQLiteRepository) CreateUser(email, name string) (int64, error) {
	var id int64
	err := r.withWrite(func() error {
		res, err := r.db.Exec(`INSERT INTO users(email,name,created_at) VALUES(?,?,?)`, email, name, time.Now().UTC().Unix())
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	return id, err
}

func (r *SQLiteRepository) GetUserByEmail(email string) (*UserRecord, error) {
	row := r.db.QueryRow(`SELECT id,email,name,created_at FROM users WHERE email=?`, email)
	var rec UserRecord
	var created int64
	if err := row.Scan(&rec.ID, &rec.Email, &rec.Name, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	rec.CreatedAt = time.Unix(created, 0).UTC()
	return &rec, nil
}

func randomAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (r *SQLiteRepository) CreateAPIKey(userID int64, scope string, botID *int64) (string, error) {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = ScopeUserAdmin
	}
	switch scope {
	case ScopeUserAdmin, ScopeUserRead:
		if botID != nil {
			return "", errors.New("bot_id not allowed for user scope")
		}
	case ScopeBotTrade:
		if botID == nil {
			return "", errors.New("bot scope requires bot_id")
		}
	default:
		return "", fmt.Errorf("unknown scope: %s", scope)
	}
	if botID != nil {
		bot, err := r.GetBot(*botID)
		if err != nil {
			return "", err
		}
		if bot.UserID != userID {
			return "", errors.New("bot owner mismatch")
		}
	}
	key, err := randomAPIKey()
	if err != nil {
		return "", err
	}
	h, err := hashKey(key)
	if err != nil {
		return "", err
	}
	enc, err := encryptAEAD([]byte(key))
	if err != nil {
		return "", err
	}
	prefix := key
	if len(prefix) > 8 {
		prefix = key[:8]
	}
	if err := r.withWrite(func() error {
		_, err = r.db.Exec(`INSERT INTO api_keys(user_id,bot_id,prefix,key_hash,secret_enc,scope,created_at) VALUES(?,?,?,?,?,?,?)`, userID, nullableInt64(botID), prefix, h, enc, scope, time.Now().UTC().Unix())
		return err
	}); err != nil {
		return "", err
	}
	return key, nil
}

func (r *SQLiteRepository) GetAPIKeyByPrefix(userID int64, prefix string) (*APIKeyRecord, error) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return nil, errors.New("empty prefix")
	}
	row := r.db.QueryRow(`SELECT id,user_id,bot_id,scope,last_used_at FROM api_keys WHERE user_id=? AND prefix=? AND revoked_at IS NULL`, userID, prefix)
	var rec APIKeyRecord
	var bot sql.NullInt64
	var lastUsed sql.NullInt64
	if err := row.Scan(&rec.ID, &rec.UserID, &bot, &rec.Scope, &lastUsed); err != nil {
		return nil, err
	}
	rec.BotID = int64PtrFromNull(bot)
	rec.LastUsedAt = timePtrFromNull(lastUsed)
	rec.Prefix = prefix
	return &rec, nil
}

func (r *SQLiteRepository) VerifyAPIKey(presented string) (*APIKeyRecord, error) {
	if strings.TrimSpace(presented) == "" {
		return nil, errors.New("empty key")
	}
	prefix := presented
	if len(prefix) > 8 {
		prefix = presented[:8]
	}
	hPresented, err := hashKey(presented)
	if err != nil {
		return nil, err
	}
	var (
		id, userID int64
		scope      string
		bot        sql.NullInt64
		lastUsed   sql.NullInt64
	)
	err = r.db.QueryRow(
		`SELECT id,user_id,bot_id,scope,last_used_at FROM api_keys WHERE prefix=? AND key_hash=? AND revoked_at IS NULL`,
		prefix,
		hPresented,
	).Scan(&id, &userID, &bot, &scope, &lastUsed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid api key")
		}
		return nil, err
	}
	now := time.Now().UTC()
	if err := r.withWrite(func() error {
		_, execErr := r.db.Exec(`UPDATE api_keys SET last_used_at=? WHERE id=?`, now.Unix(), id)
		return execErr
	}); err != nil {
		return nil, err
	}
	rec := &APIKeyRecord{
		ID:         id,
		UserID:     userID,
		BotID:      int64PtrFromNull(bot),
		Scope:      scope,
		Prefix:     prefix,
		LastUsedAt: &now,
	}
	return rec, nil
}

func (r *SQLiteRepository) RotateAPIKey(userID int64, oldPrefix string) (string, error) {
	var scope = ScopeUserAdmin
	var botPtr *int64
	oldPrefix = strings.TrimSpace(oldPrefix)
	if oldPrefix != "" {
		existing, err := r.GetAPIKeyByPrefix(userID, oldPrefix)
		if err == nil {
			scope = existing.Scope
			botPtr = existing.BotID
			if err := r.withWrite(func() error {
				_, err := r.db.Exec(`UPDATE api_keys SET revoked_at=? WHERE id=?`, time.Now().UTC().Unix(), existing.ID)
				return err
			}); err != nil {
				return "", err
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	return r.CreateAPIKey(userID, scope, botPtr)
}

func (r *SQLiteRepository) RevokeAPIKey(userID int64, prefix string) error {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return errors.New("empty prefix")
	}
	var res sql.Result
	if err := r.withWrite(func() error {
		var err error
		res, err = r.db.Exec(`UPDATE api_keys SET revoked_at=? WHERE user_id=? AND prefix=? AND revoked_at IS NULL`, time.Now().UTC().Unix(), userID, prefix)
		return err
	}); err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SQLiteRepository) StoreWallet(userID int64, address string, privPlain []byte) error {
	enc, err := encryptAEAD(privPlain)
	if err != nil {
		return err
	}
	return r.withWrite(func() error {
		_, err = r.db.Exec(`INSERT INTO wallets(user_id,address,priv_enc,created_at) VALUES(?,?,?,?) ON CONFLICT(address) DO UPDATE SET priv_enc=excluded.priv_enc`, userID, address, enc, time.Now().UTC().Unix())
		return err
	})
}

func (r *SQLiteRepository) LoadWallet(userID int64) (string, []byte, error) {
	row := r.db.QueryRow(`SELECT address,priv_enc FROM wallets WHERE user_id=? ORDER BY id DESC LIMIT 1`, userID)
	var addr string
	var enc []byte
	if err := row.Scan(&addr, &enc); err != nil {
		return "", nil, err
	}
	plain, err := decryptAEAD(enc)
	if err != nil {
		return "", nil, err
	}
	return addr, plain, nil
}

func (r *SQLiteRepository) CreateBot(userID int64, name string, configJSON string) (int64, error) {
	var id int64
	err := r.withWrite(func() error {
		res, err := r.db.Exec(`INSERT INTO bots(user_id,name,config_json,created_at) VALUES(?,?,?,?)`, userID, name, configJSON, time.Now().UTC().Unix())
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	return id, err
}

func (r *SQLiteRepository) GetBot(botID int64) (*BotRecord, error) {
	row := r.db.QueryRow(`SELECT id,user_id,name,config_json,created_at FROM bots WHERE id=?`, botID)
	var rec BotRecord
	var created int64
	if err := row.Scan(&rec.ID, &rec.UserID, &rec.Name, &rec.ConfigJSON, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	rec.CreatedAt = time.Unix(created, 0).UTC()
	return &rec, nil
}

func (r *SQLiteRepository) ListBots(userID int64) ([]BotRecord, error) {
	rows, err := r.db.Query(`SELECT id,user_id,name,config_json,created_at FROM bots WHERE user_id=? ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]BotRecord, 0, 8)
	for rows.Next() {
		var rec BotRecord
		var created int64
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.Name, &rec.ConfigJSON, &created); err != nil {
			return nil, err
		}
		rec.CreatedAt = time.Unix(created, 0).UTC()
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *SQLiteRepository) GetBotByName(name string) (*BotRecord, error) {
	row := r.db.QueryRow(`SELECT id,user_id,name,config_json,created_at FROM bots WHERE name=? ORDER BY created_at DESC LIMIT 1`, name)
	var rec BotRecord
	var created int64
	if err := row.Scan(&rec.ID, &rec.UserID, &rec.Name, &rec.ConfigJSON, &created); err != nil {
		return nil, err
	}
	rec.CreatedAt = time.Unix(created, 0).UTC()
	return &rec, nil
}

func (r *SQLiteRepository) UpdateBotConfig(botID int64, configJSON string) error {
	return r.withWrite(func() error {
		_, err := r.db.Exec(`UPDATE bots SET config_json=? WHERE id=?`, configJSON, botID)
		return err
	})
}

func (r *SQLiteRepository) CreateGrant(botID int64, grantType string, paramsJSON string, expiresAt *time.Time) (int64, error) {
	var exp *int64
	if expiresAt != nil {
		v := expiresAt.UTC().Unix()
		exp = &v
	}
	var id int64
	err := r.withWrite(func() error {
		res, err := r.db.Exec(`INSERT INTO grants(bot_id,grant_type,params_json,created_at,expires_at,revoked_at) VALUES(?,?,?,?,?,?)`, botID, grantType, paramsJSON, time.Now().UTC().Unix(), exp, nil)
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	return id, err
}

func (r *SQLiteRepository) LogAudit(userID *int64, botID *int64, apiKeyID *int64, action string, metaJSON string) error {
	action = strings.TrimSpace(action)
	if action == "" {
		return errors.New("action required")
	}
	if strings.TrimSpace(metaJSON) == "" {
		metaJSON = "{}"
	}
	return r.withWrite(func() error {
		_, err := r.db.Exec(`INSERT INTO audits(user_id,bot_id,api_key_id,action,meta_json,created_at) VALUES(?,?,?,?,?,?)`, nullableInt64(userID), nullableInt64(botID), nullableInt64(apiKeyID), action, metaJSON, time.Now().UTC().Unix())
		return err
	})
}

func (r *SQLiteRepository) ListAudits(userID int64, limit int) ([]AuditRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id,user_id,bot_id,api_key_id,action,meta_json,created_at FROM audits WHERE user_id=? ORDER BY created_at DESC, id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AuditRecord, 0, limit)
	for rows.Next() {
		var rec AuditRecord
		var ts int64
		var user sql.NullInt64
		var bot sql.NullInt64
		var apiKey sql.NullInt64
		if err := rows.Scan(&rec.ID, &user, &bot, &apiKey, &rec.Action, &rec.MetaJSON, &ts); err != nil {
			return nil, err
		}
		rec.UserID = int64PtrFromNull(user)
		rec.BotID = int64PtrFromNull(bot)
		rec.APIKeyID = int64PtrFromNull(apiKey)
		rec.CreatedAt = time.Unix(ts, 0).UTC()
		out = append(out, rec)
	}
	return out, rows.Err()
}
