package signerMock

import (
	"crypto/cipher"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type signerMock struct {
	ethereumAddress            func() common.Address
	getSharedKey               func(their ecdsa.PublicKey) [32]byte
	genNonce                   func() string
	encryptAndGetChallengeHash func(key string, message string) (string, string, error)
	decryptMessage             func(sharedKey string, cipherText string) (string, error)
	sign                       func(hash string) (string, error)
	getPublicKey               func() *ecdsa.PublicKey
	publicKeyFromBytes         func(pbKey []byte) (*ecdsa.PublicKey, error)
	bytesFromPublicKey         func(key *ecdsa.PublicKey) []byte
	getPrivateKey              func() *ecdsa.PrivateKey
	getCipherMode              func(key string) (cipher.AEAD, error)
	verifySignature            func(publicKey ecdsa.PublicKey, signature, messageHash string) bool
	signTx                     func(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error)
}

func (s *signerMock) EthereumAddress() common.Address {
	if s.ethereumAddress != nil {
		return s.ethereumAddress()
	}
	return common.Address{}
}

func (s *signerMock) GetSharedKey(their ecdsa.PublicKey) [32]byte {
	if s.getSharedKey != nil {
		return s.getSharedKey(their)
	}
	return [32]byte{}
}

func (s *signerMock) GenNonce() string {
	if s.genNonce != nil {
		return s.genNonce()
	}
	return ""
}

func (s *signerMock) EncryptAndGetChallengeHash(key string, nonce string, message string) (string, string, error) {
	if s.encryptAndGetChallengeHash != nil {
		return key, nonce, nil
	}
	return "", "", errors.New("EncryptAndGetHash not implemented")
}

func (s *signerMock) DecryptMessage(sharedKey string, cipherText string) (string, error) {
	if s.decryptMessage != nil {
		return s.decryptMessage(sharedKey, cipherText)
	}
	return "", errors.New("DecryptMessage not implemented")
}

func (s *signerMock) GetCipherMode(key string) (cipher.AEAD, error) {
	if s.getCipherMode != nil {
		return s.getCipherMode(key)
	}
	return nil, errors.New("getCipherMode not implemented")
}

func (s *signerMock) VerifySignature(publicKey ecdsa.PublicKey, signature, signedHash string) bool {
	if s.verifySignature != nil {
		return s.verifySignature(publicKey, signedHash, signedHash)
	}
	return false
}

func (s *signerMock) Sign(hash string) (string, error) {
	if s.sign != nil {
		return s.sign(hash)
	}
	return "", errors.New("sign not implemented")
}

func (s *signerMock) GetPublicKey() *ecdsa.PublicKey {
	if s.getPublicKey != nil {
		return s.getPublicKey()
	}
	return nil
}

func (s *signerMock) PublicKeyFromBytes(pbKey []byte) (*ecdsa.PublicKey, error) {
	if s.publicKeyFromBytes != nil {
		return s.publicKeyFromBytes(pbKey)
	}
	return nil, errors.New("PublicKeyFromBytes not implemented")
}

func (s *signerMock) BytesFromPublicKey(key *ecdsa.PublicKey) []byte {
	if s.bytesFromPublicKey != nil {
		return s.bytesFromPublicKey(key)
	}
	return nil
}

func (s *signerMock) GetPrivateKey() *ecdsa.PrivateKey {
	if s.getPrivateKey != nil {
		return s.getPrivateKey()
	}
	return nil
}

// SignTx signs an ethereum transaction.
func (s *signerMock) SignTx(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	txSigner := types.NewLondonSigner(chainID)

	if s.getPrivateKey == nil {
		return nil, errors.New("getPrivateKey function is not set")
	}

	privateKey := s.getPrivateKey()
	if privateKey == nil {
		return nil, errors.New("no private key provided")
	}

	signedTx, err := types.SignTx(transaction, txSigner, privateKey)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

// Option is the option passed to the mock service
type Option interface {
	apply(mock *signerMock)
}

type optionFunc func(mock *signerMock)

func (f optionFunc) apply(r *signerMock) { f(r) }

func New(opts ...Option) *signerMock {
	mock := new(signerMock)
	for _, o := range opts {
		o.apply(mock)
	}
	return mock
}

// WithEthereumAddressFunc sets the ethereumAddress function.
func WithEthereumAddressFunc(f func() common.Address) Option {
	return optionFunc(func(sm *signerMock) {
		sm.ethereumAddress = f
	})
}

// WithSignTxFunc sets the signTx function.
func WithSignTxFunc(f func(*types.Transaction, *big.Int) (*types.Transaction, error)) Option {
	return optionFunc(func(sm *signerMock) {
		sm.signTx = f
	})
}

// Repeat the pattern for all other methods:

func WithGetSharedKeyFunc(f func(their ecdsa.PublicKey) [32]byte) Option {
	return optionFunc(func(sm *signerMock) {
		sm.getSharedKey = f
	})
}

func WithGenNonceFunc(f func() string) Option {
	return optionFunc(func(sm *signerMock) {
		sm.genNonce = f
	})
}

func WithEncryptAndGetChallengeHash(f func(key string, msg string) (string, string, error)) Option {
	return optionFunc(func(sm *signerMock) {
		sm.encryptAndGetChallengeHash = f
	})
}

func WithDecryptMessageFunc(f func(string, string) (string, error)) Option {
	return optionFunc(func(sm *signerMock) {
		sm.decryptMessage = f
	})
}

func WithCipherModeFunc(f func(key string) (cipher.AEAD, error)) Option {
	return optionFunc(func(sm *signerMock) {
		sm.getCipherMode = f
	})
}

func WithVerifySignatureFunc(f func(ecdsa.PublicKey, string, string) bool) Option {
	return optionFunc(func(sm *signerMock) {
		sm.verifySignature = f
	})
}

func WithSignFunc(f func(hash string) (string, error)) Option {
	return optionFunc(func(sm *signerMock) {
		sm.sign = f
	})
}

func WithGetPublicKeyFunc(f func() *ecdsa.PublicKey) Option {
	return optionFunc(func(sm *signerMock) {
		sm.getPublicKey = f
	})
}

func WithPublicKeyFromBytesFunc(f func([]byte) (*ecdsa.PublicKey, error)) Option {
	return optionFunc(func(sm *signerMock) {
		sm.publicKeyFromBytes = f
	})
}

func WithBytesFromPublicKeyFunc(f func(*ecdsa.PublicKey) []byte) Option {
	return optionFunc(func(sm *signerMock) {
		sm.bytesFromPublicKey = f
	})
}

func WithGetPrivateKeyFunc(f func() *ecdsa.PrivateKey) Option {
	return optionFunc(func(sm *signerMock) {
		sm.getPrivateKey = f
	})
}
