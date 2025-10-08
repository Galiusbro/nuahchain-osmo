package types

import "fmt"

// DefaultGenesis returns default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Tokens: []Token{},
	}
}

// Validate performs basic genesis state validation.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seenDenoms := make(map[string]struct{})
	seenNames := make(map[string]struct{})
	seenSymbols := make(map[string]struct{})

	for _, token := range gs.Tokens {
		if err := ValidateTokenBasic(token); err != nil {
			return err
		}

		if _, exists := seenDenoms[token.Denom]; exists {
			return fmt.Errorf("duplicate token denom: %s", token.Denom)
		}
		seenDenoms[token.Denom] = struct{}{}

		nameKey := normalizeName(token.Name)
		if _, exists := seenNames[nameKey]; exists {
			return fmt.Errorf("duplicate token name: %s", token.Name)
		}
		seenNames[nameKey] = struct{}{}

		symbolKey := normalizeSymbol(token.Symbol)
		if _, exists := seenSymbols[symbolKey]; exists {
			return fmt.Errorf("duplicate token symbol: %s", token.Symbol)
		}
		seenSymbols[symbolKey] = struct{}{}
	}

	return nil
}
