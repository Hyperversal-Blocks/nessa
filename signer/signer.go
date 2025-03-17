package signer

import (
	"crypto/cipher"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type signer struct {
	PublicKey  *ecdsa.PublicKey
	PrivateKey *ecdsa.PrivateKey
	Address    *string
}

type Signer interface {
	// EthereumAddress returns the ethereum address this signer uses.
	EthereumAddress() common.Address
	SignTx(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error)
	GetSharedKey(their ecdsa.PublicKey) string
	GenNonce() string
	EncryptAndGetChallengeHash(key string, message string) (string, string, error)
	DecryptMessage(sharedKey string, cipherText string) (string, error)
	GetCipherMode(key string) (cipher.AEAD, error)
	VerifySignature(publicKey ecdsa.PublicKey, signature, messageHash string) bool
	Sign(hash string) (string, error)
	GetPublicKey() *ecdsa.PublicKey
	PublicKeyFromBytes(pbKey string) (*ecdsa.PublicKey, error)
	BytesFromPublicKey(key *ecdsa.PublicKey) []byte
	GetPrivateKey() *ecdsa.PrivateKey
}

func (signerObject *signer) GetPublicKey() *ecdsa.PublicKey {
	return signerObject.PublicKey
}

func (signerObject *signer) EthereumAddress() common.Address {
	return crypto.PubkeyToAddress(*signerObject.PublicKey)
}

// SignTx signs an ethereum transaction.
func (signerObject *signer) SignTx(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	txSigner := types.NewLondonSigner(chainID)

	signedTx, err := types.SignTx(transaction, txSigner, signerObject.PrivateKey)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func (signerObject *signer) GetPrivateKey() *ecdsa.PrivateKey {
	return signerObject.PrivateKey
}

func New() (Signer, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create a signer: %w", err)
	}

	publicKey := privateKey.Public()

	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to create a child from signer: %w", err)
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return &signer{
		PublicKey:  publicKeyECDSA,
		PrivateKey: privateKey,
		Address:    &address,
	}, nil
}

func NewFromKey(hexedPvtKey string) (Signer, error) {
	privateKey, err := crypto.HexToECDSA(hexedPvtKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create a signer: %w", err)
	}

	publicKey := privateKey.Public()

	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to create a child from signer: %w", err)
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return &signer{
		PublicKey:  publicKeyECDSA,
		PrivateKey: privateKey,
		Address:    &address,
	}, nil
}
