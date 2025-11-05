package auth

import (
	"database/sql"
	"errors"
	"time"
)

// Repository handles database operations for authentication
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser creates a new user
func (r *Repository) CreateUser(email, username, passwordHash *string, telegramID *int64, telegramUsername *string) (*User, error) {
	var userID int64
	err := r.db.QueryRow(`
		INSERT INTO users (email, username, password_hash, telegram_id, telegram_username, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`, email, username, passwordHash, telegramID, telegramUsername).Scan(&userID)
	if err != nil {
		return nil, err
	}

	return r.GetUserByID(userID)
}

// GetUserByID gets a user by ID
func (r *Repository) GetUserByID(id int64) (*User, error) {
	user := &User{ID: id}
	err := r.db.QueryRow(`
		SELECT email, username, telegram_id, telegram_username, created_at, updated_at, last_login_at, is_active
		FROM users WHERE id = $1
	`, id).Scan(
		&user.Email, &user.Username, &user.TelegramID, &user.TelegramUsername,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// GetUserByEmail gets a user by email
func (r *Repository) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	var passwordHash sql.NullString
	err := r.db.QueryRow(`
		SELECT id, email, username, password_hash, telegram_id, telegram_username,
		       created_at, updated_at, last_login_at, is_active
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.Username, &passwordHash, &user.TelegramID, &user.TelegramUsername,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// GetUserByEmailWithPassword gets a user by email with password hash
func (r *Repository) GetUserByEmailWithPassword(email string) (*User, string, error) {
	user := &User{}
	var passwordHash sql.NullString
	err := r.db.QueryRow(`
		SELECT id, email, username, password_hash, telegram_id, telegram_username,
		       created_at, updated_at, last_login_at, is_active
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.Username, &passwordHash, &user.TelegramID, &user.TelegramUsername,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("user not found")
		}
		return nil, "", err
	}

	if !passwordHash.Valid {
		return nil, "", errors.New("user has no password")
	}

	return user, passwordHash.String, nil
}

// GetUserByTelegramID gets a user by Telegram ID
func (r *Repository) GetUserByTelegramID(telegramID int64) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(`
		SELECT id, email, username, telegram_id, telegram_username,
		       created_at, updated_at, last_login_at, is_active
		FROM users WHERE telegram_id = $1
	`, telegramID).Scan(
		&user.ID, &user.Email, &user.Username, &user.TelegramID, &user.TelegramUsername,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// UpdateUserLastLogin updates the user's last login time
func (r *Repository) UpdateUserLastLogin(userID int64) error {
	_, err := r.db.Exec(`
		UPDATE users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1
	`, userID)
	return err
}

// CreateWallet creates a new wallet for a user
func (r *Repository) CreateWallet(userID int64, address string, encryptedPrivateKey, encryptedMnemonic []byte) (*Wallet, error) {
	var walletID int64
	err := r.db.QueryRow(`
		INSERT INTO wallets (user_id, address, encrypted_private_key, mnemonic_encrypted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, userID, address, encryptedPrivateKey, encryptedMnemonic).Scan(&walletID)
	if err != nil {
		return nil, err
	}

	return r.GetWalletByID(walletID)
}

// GetWalletByUserID gets a wallet by user ID
func (r *Repository) GetWalletByUserID(userID int64) (*Wallet, error) {
	wallet := &Wallet{UserID: userID}
	err := r.db.QueryRow(`
		SELECT id, address, encrypted_private_key, mnemonic_encrypted, created_at, updated_at
		FROM wallets WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(
		&wallet.ID, &wallet.Address, &wallet.EncryptedPrivateKey,
		&wallet.MnemonicEncrypted, &wallet.CreatedAt, &wallet.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}
	return wallet, nil
}

// GetWalletByID gets a wallet by ID
func (r *Repository) GetWalletByID(id int64) (*Wallet, error) {
	wallet := &Wallet{ID: id}
	err := r.db.QueryRow(`
		SELECT user_id, address, encrypted_private_key, mnemonic_encrypted, created_at, updated_at
		FROM wallets WHERE id = $1
	`, id).Scan(
		&wallet.UserID, &wallet.Address, &wallet.EncryptedPrivateKey,
		&wallet.MnemonicEncrypted, &wallet.CreatedAt, &wallet.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}
	return wallet, nil
}

// GetWalletByAddress gets a wallet by address
func (r *Repository) GetWalletByAddress(address string) (*Wallet, error) {
	wallet := &Wallet{}
	err := r.db.QueryRow(`
		SELECT id, user_id, address, encrypted_private_key, mnemonic_encrypted, created_at, updated_at
		FROM wallets WHERE address = $1
	`, address).Scan(
		&wallet.ID, &wallet.UserID, &wallet.Address, &wallet.EncryptedPrivateKey,
		&wallet.MnemonicEncrypted, &wallet.CreatedAt, &wallet.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}
	return wallet, nil
}

// CreateSession creates a new session
func (r *Repository) CreateSession(userID int64, token, refreshToken string, expiresAt, refreshExpiresAt time.Time, ipAddress, userAgent *string) (*Session, error) {
	var sessionID int64
	err := r.db.QueryRow(`
		INSERT INTO sessions (user_id, token, refresh_token, expires_at, refresh_expires_at, ip_address, user_agent, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`, userID, token, refreshToken, expiresAt, refreshExpiresAt, ipAddress, userAgent).Scan(&sessionID)
	if err != nil {
		return nil, err
	}

	return r.GetSessionByToken(token)
}

// GetSessionByToken gets a session by token
func (r *Repository) GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	err := r.db.QueryRow(`
		SELECT id, user_id, token, refresh_token, expires_at, refresh_expires_at,
		       ip_address, user_agent, created_at, last_used_at, revoked_at
		FROM sessions WHERE token = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.RefreshToken,
		&session.ExpiresAt, &session.RefreshExpiresAt, &session.IPAddress, &session.UserAgent,
		&session.CreatedAt, &session.LastUsedAt, &session.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found or expired")
		}
		return nil, err
	}
	return session, nil
}

// RevokeSession revokes a session
func (r *Repository) RevokeSession(token string) error {
	_, err := r.db.Exec(`
		UPDATE sessions SET revoked_at = NOW() WHERE token = $1
	`, token)
	return err
}

// UpdateSessionLastUsed updates the session's last used time
func (r *Repository) UpdateSessionLastUsed(token string) error {
	_, err := r.db.Exec(`
		UPDATE sessions SET last_used_at = NOW() WHERE token = $1
	`, token)
	return err
}
