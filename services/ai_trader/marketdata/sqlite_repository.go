package marketdata

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

// Repository abstracts persistence for prices and candles.
type Repository interface {
	SaveLatest(p Price) error
	AppendCandles(symbol string, tf Timeframe, bars []Candle) error
	GetCandles(symbol string, tf Timeframe, from, to time.Time, limit int) ([]Candle, error)
}

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dsn string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;`); err != nil {
		return nil, err
	}
	r := &SQLiteRepository{db: db}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *SQLiteRepository) migrate() error {
	_, err := r.db.Exec(`
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
`)
	return err
}

func (r *SQLiteRepository) SaveLatest(p Price) error {
	_, err := r.db.Exec(`INSERT INTO prices(symbol,value,source,ts_unix) VALUES(?,?,?,?)
ON CONFLICT(symbol) DO UPDATE SET value=excluded.value, source=excluded.source, ts_unix=excluded.ts_unix`, p.Symbol, p.Value, p.Source, p.Timestamp.Unix())
	return err
}

func (r *SQLiteRepository) AppendCandles(symbol string, tf Timeframe, bars []Candle) error {
	if len(bars) == 0 {
		return nil
	}
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
