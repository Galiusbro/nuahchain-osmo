package tokens

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// Repository handles database operations for tokens
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new tokens repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Token represents a token with metadata
type Token struct {
	ID             int64     `json:"id"`
	Denom          string    `json:"denom"`
	Name           string    `json:"name"`
	Symbol         string    `json:"symbol"`
	Image          *string   `json:"image,omitempty"`
	Description    *string   `json:"description,omitempty"`
	CreatorAddress string    `json:"creator_address"`
	CreatorUserID  *int64    `json:"creator_user_id,omitempty"`
	Decimals       int       `json:"decimals"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateOrUpdateToken creates a new token or updates existing one
func (r *Repository) CreateOrUpdateToken(token Token) error {
	_, err := r.db.Exec(`
		INSERT INTO tokens (
			denom,
			name,
			symbol,
			image,
			description,
			creator_address,
			creator_user_id,
			decimals,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (denom) DO UPDATE SET
			name = EXCLUDED.name,
			symbol = EXCLUDED.symbol,
			image = EXCLUDED.image,
			description = EXCLUDED.description,
			updated_at = NOW()
	`, token.Denom, token.Name, token.Symbol, token.Image, token.Description, token.CreatorAddress, token.CreatorUserID, token.Decimals)
	return err
}

// GetTokenByDenom gets token metadata by denom
func (r *Repository) GetTokenByDenom(denom string) (*Token, error) {
	token := &Token{}
	var image, description sql.NullString
	var creatorUserID sql.NullInt64

	err := r.db.QueryRow(`
		SELECT
			id,
			denom,
			name,
			symbol,
			image,
			description,
			creator_address,
			creator_user_id,
			decimals,
			created_at,
			updated_at
		FROM tokens
		WHERE denom = $1
	`, denom).Scan(
		&token.ID,
		&token.Denom,
		&token.Name,
		&token.Symbol,
		&image,
		&description,
		&token.CreatorAddress,
		&creatorUserID,
		&token.Decimals,
		&token.CreatedAt,
		&token.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Token not found, return nil (not an error)
		}
		return nil, err
	}

	if image.Valid {
		token.Image = &image.String
	}
	if description.Valid {
		token.Description = &description.String
	}
	if creatorUserID.Valid {
		token.CreatorUserID = &creatorUserID.Int64
	}

	return token, nil
}

// GetTokensByDenoms gets multiple tokens by their denoms
func (r *Repository) GetTokensByDenoms(denoms []string) (map[string]*Token, error) {
	if len(denoms) == 0 {
		return make(map[string]*Token), nil
	}

	// Build query with placeholders
	query := `
		SELECT
			id,
			denom,
			name,
			symbol,
			image,
			description,
			creator_address,
			creator_user_id,
			decimals,
			created_at,
			updated_at
		FROM tokens
		WHERE denom = ANY($1)
	`

	rows, err := r.db.Query(query, pq.Array(denoms))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make(map[string]*Token)
	for rows.Next() {
		token := &Token{}
		var image, description sql.NullString
		var creatorUserID sql.NullInt64

		err := rows.Scan(
			&token.ID,
			&token.Denom,
			&token.Name,
			&token.Symbol,
			&image,
			&description,
			&token.CreatorAddress,
			&creatorUserID,
			&token.Decimals,
			&token.CreatedAt,
			&token.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if image.Valid {
			token.Image = &image.String
		}
		if description.Valid {
			token.Description = &description.String
		}
		if creatorUserID.Valid {
			token.CreatorUserID = &creatorUserID.Int64
		}

		tokens[token.Denom] = token
	}

	return tokens, rows.Err()
}

