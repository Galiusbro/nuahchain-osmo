package auth

import (
	"crypto/rand"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

// WalletGenerator handles wallet creation
type WalletGenerator struct{}

// NewWalletGenerator creates a new wallet generator
func NewWalletGenerator() *WalletGenerator {
	return &WalletGenerator{}
}

// GenerateWallet generates a new Cosmos wallet with mnemonic
func (wg *WalletGenerator) GenerateWallet() (address string, privateKey []byte, mnemonic string, err error) {
	// Generate entropy for mnemonic
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	// Generate mnemonic
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	// Derive private key from mnemonic
	seed := bip39.NewSeed(mnemonic, "")
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, "44'/118'/0'/0/0")
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to derive private key: %w", err)
	}

	// Create secp256k1 private key
	privKey := &cosmossecp256k1.PrivKey{Key: derivedPriv}

	// Get address
	pubKey := privKey.PubKey()
	address = sdk.AccAddress(pubKey.Address()).String()

	return address, privKey.Key, mnemonic, nil
}

// GenerateWalletFromMnemonic generates a wallet from existing mnemonic
func (wg *WalletGenerator) GenerateWalletFromMnemonic(mnemonic string) (address string, privateKey []byte, err error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", nil, fmt.Errorf("invalid mnemonic")
	}

	// Derive private key from mnemonic
	seed := bip39.NewSeed(mnemonic, "")
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, "44'/118'/0'/0/0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	// Create secp256k1 private key
	privKey := &secp256k1.PrivKey{Key: derivedPriv}

	// Get address
	pubKey := privKey.PubKey()
	address = sdk.AccAddress(pubKey.Address()).String()

	return address, privKey.Key, nil
}

// GenerateRandomBytes generates random bytes of specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
