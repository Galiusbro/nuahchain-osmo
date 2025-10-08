package types

import (
	"net/url"
	"regexp"
	"strings"
)

const (
	MinNameLength        = 3
	MaxNameLength        = 64
	MinSymbolLength      = 3
	MaxSymbolLength      = 16
	MaxDescriptionLength = 1024
	MaxImageLength       = 512
)

var (
	symbolRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
)

// ValidateTokenBasic performs stateless validation on a token definition.
func ValidateTokenBasic(token Token) error {
	if err := ValidateName(token.Name); err != nil {
		return err
	}
	if err := ValidateSymbol(token.Symbol); err != nil {
		return err
	}
	if err := ValidateImage(token.Image); err != nil {
		return err
	}
	if err := ValidateDescription(token.Description); err != nil {
		return err
	}
	return nil
}

// ValidateName ensures a token name meets formatting constraints.
func ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < MinNameLength {
		return ErrInvalidName
	}
	if len(trimmed) > MaxNameLength {
		return ErrInvalidName
	}
	return nil
}

// ValidateSymbol ensures a token symbol meets formatting constraints.
func ValidateSymbol(symbol string) error {
	trimmed := strings.TrimSpace(symbol)
	if len(trimmed) < MinSymbolLength || len(trimmed) > MaxSymbolLength {
		return ErrInvalidSymbol
	}
	if !symbolRegexp.MatchString(trimmed) {
		return ErrInvalidSymbol
	}
	return nil
}

// ValidateImage performs a lightweight validation of the image URL.
func ValidateImage(image string) error {
	trimmed := strings.TrimSpace(image)
	if trimmed == "" {
		return nil
	}
	if len(trimmed) > MaxImageLength {
		return ErrInvalidImage
	}
	u, err := url.Parse(trimmed)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ErrInvalidImage
	}
	return nil
}

// ValidateDescription ensures descriptions stay within a bounded size.
func ValidateDescription(desc string) error {
	if len(desc) > MaxDescriptionLength {
		return ErrInvalidDescription
	}
	return nil
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizeSymbol(symbol string) string {
	return strings.ToLower(strings.TrimSpace(symbol))
}
