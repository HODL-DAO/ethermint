package crypto

import (
	"fmt"

	"github.com/pkg/errors"

	"crypto/hmac"
	"crypto/sha512"

	"github.com/tyler-smith/go-bip39"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

const (
	// EthSecp256k1 defines the ECDSA secp256k1 used on Ethereum
	EthSecp256k1 = keyring.SigningAlgo("eth_secp256k1")
)

// SupportedAlgorithms defines the list of signing algorithms used on Ethermint:
//  - eth_secp256k1 (Ethereum)
//  - secp256k1 (Tendermint)
var SupportedAlgorithms = []keyring.SigningAlgo{EthSecp256k1, keyring.Secp256k1}

// EthSecp256k1Options defines a keyring options for the ethereum Secp256k1 curve.
func EthSecp256k1Options() []keyring.KeybaseOption {
	return []keyring.KeybaseOption{
		keyring.WithKeygenFunc(EthermintKeygenFunc),
		keyring.WithDeriveFunc(DeriveKey),
		keyring.WithSupportedAlgos(SupportedAlgorithms),
		keyring.WithSupportedAlgosLedger(SupportedAlgorithms),
	}
}

func DeriveKey(mnemonic, bip39Passphrase, hdPath string, algo keyring.SigningAlgo) ([]byte, error) {
	switch algo {
	case keyring.Secp256k1:
		return keyring.StdDeriveKey(mnemonic, bip39Passphrase, hdPath, algo)
	case EthSecp256k1:
		return DeriveSecp256k1(mnemonic, bip39Passphrase, hdPath)
	default:
		return nil, errors.Wrap(keyring.ErrUnsupportedSigningAlgo, string(algo))
	}
}

// EthermintKeygenFunc is the key generation function to generate secp256k1 ToECDSA
// from ethereum.
func EthermintKeygenFunc(bz []byte, algo keyring.SigningAlgo) (tmcrypto.PrivKey, error) {
	if algo != EthSecp256k1 {
		return nil, fmt.Errorf("signing algorithm must be %s, got %s", EthSecp256k1, algo)
	}

	return PrivKeySecp256k1(bz), nil
}

func DeriveSecp256k1(mnemonic, bip39Passphrase, _ string) ([]byte, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return nil, err
	}

	// HMAC the seed to produce the private key and chain code
	mac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	_, err = mac.Write(seed)
	if err != nil {
		return nil, err
	}

	seed = mac.Sum(nil)

	priv, err := ethcrypto.ToECDSA(seed[:32])
	if err != nil {
		return nil, err
	}

	derivedKey := PrivKeySecp256k1(ethcrypto.FromECDSA(priv))

	return derivedKey, nil
}
