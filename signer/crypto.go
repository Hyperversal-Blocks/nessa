package signer

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

// GetSharedKey returns the shared key using the private and public key.
func (signerObject *signer) GetSharedKey(their ecdsa.PublicKey) string {
	sharedKey, _ := their.Curve.ScalarMult(their.X, their.Y, signerObject.PrivateKey.D.Bytes())
	return hex.EncodeToString(sharedKey.Bytes())
}

// GenNonce generates a nonce for AES-GCM (12 bytes) and returns it as a hex string.
func (signerObject *signer) GenNonce() string {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return ""
	}
	return hex.EncodeToString(nonce)
}

func (signerObject *signer) EncryptAndGetChallengeHash(key string, message string) (string, string, error) {
	aesgcm, err := signerObject.GetCipherMode(key)
	if err != nil {
		return "", "", fmt.Errorf("error getting cipher mode: %w", err)
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(message), nil)

	combined := append(nonce, ciphertext...)

	hasher := sha3.New256()
	if _, err := hasher.Write([]byte(key)); err != nil {
		return "", "", fmt.Errorf("error hashing key: %w", err)
	}

	hashedEphemeralKey := hasher.Sum(nil)

	return hex.EncodeToString(hashedEphemeralKey), hex.EncodeToString(combined), nil
}

// DecryptMessage using sharedKey, ciphered text and the nonce used to encrypt it.
func (signerObject *signer) DecryptMessage(sharedKey string, combinedHex string) (string, error) {
	aesgcm, err := signerObject.GetCipherMode(sharedKey)
	if err != nil {
		return "", fmt.Errorf("error getting cipher mode: %w", err)
	}

	combinedBytes, err := hex.DecodeString(combinedHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode combined ciphertext: %w", err)
	}
	if len(combinedBytes) < aesgcm.NonceSize() {
		return "", fmt.Errorf("combined data too short for nonce")
	}

	nonce := combinedBytes[:aesgcm.NonceSize()]
	ciphertext := combinedBytes[aesgcm.NonceSize():]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("error decrypting: %w", err)
	}

	return string(plaintext), nil
}

func (signerObject *signer) GetCipherMode(key string) (cipher.AEAD, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from hex: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode with nonce size 32: %w", err)
	}

	return aesgcm, nil
}

func (signerObject *signer) VerifySignature(publicKey ecdsa.PublicKey, challenge, signatureHex string) bool {
	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}
	if len(sigBytes) != 64 {
		return false
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	sVal := new(big.Int).SetBytes(sigBytes[32:])

	hasher := sha3.New256()
	hasher.Write([]byte(challenge))
	hashBytes := hasher.Sum(nil)

	// Verify the signature.
	return ecdsa.Verify(&publicKey, hashBytes, r, sVal)
}

// Sign the hash with privateKey of encrypter.
func (signerObject *signer) Sign(hash string) (string, error) {
	r, s, err := ecdsa.Sign(rand.Reader, signerObject.PrivateKey, []byte(hash[:]))
	if err != nil {
		return "", fmt.Errorf("error signing using private key: %w", err)
	}

	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature), nil
}

func (signerObject *signer) PublicKeyFromBytes(pbKey string) (*ecdsa.PublicKey, error) {
	keyBytes, err := hex.DecodeString(pbKey)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPubkey(keyBytes)
}

func (signerObject *signer) BytesFromPublicKey(key *ecdsa.PublicKey) []byte {
	return crypto.FromECDSAPub(key)
}
