package auth

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID               int64      `json:"id"`
	Email            *string    `json:"email,omitempty"`
	Username         *string    `json:"username,omitempty"`
	TelegramID       *int64     `json:"telegram_id,omitempty"`
	TelegramUsername *string    `json:"telegram_username,omitempty"`
	ImageURL         *string    `json:"image_url,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	IsActive         bool       `json:"is_active"`
}

// Wallet represents a user's blockchain wallet
type Wallet struct {
	ID                  int64     `json:"id"`
	UserID              int64     `json:"user_id"`
	Address             string    `json:"address"`
	EncryptedPrivateKey []byte    `json:"-"` // Never expose in JSON
	MnemonicEncrypted   []byte    `json:"-"` // Never expose in JSON
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	Token            string     `json:"token"`
	RefreshToken     *string    `json:"refresh_token,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RefreshExpiresAt *time.Time `json:"refresh_expires_at,omitempty"`
	IPAddress        *string    `json:"ip_address,omitempty"`
	UserAgent        *string    `json:"user_agent,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	LastUsedAt       time.Time  `json:"last_used_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
}

// TelegramAuthData represents Telegram authentication data
type TelegramAuthData struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	TelegramID int64     `json:"telegram_id"`
	FirstName  *string   `json:"first_name,omitempty"`
	LastName   *string   `json:"last_name,omitempty"`
	Username   *string   `json:"username,omitempty"`
	AuthDate   int64     `json:"auth_date"`
	Hash       string    `json:"hash"`
	CreatedAt  time.Time `json:"created_at"`
}
