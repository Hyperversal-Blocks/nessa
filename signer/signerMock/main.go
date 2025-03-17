package signerMock

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

// ComputeSharedKey computes the shared key using ECDH with the server's public key and the client's private key.
func ComputeSharedKey(serverPubKey *ecdsa.PublicKey, clientPrivKey *ecdsa.PrivateKey) (string, error) {
	// Perform ECDH scalar multiplication.
	sharedX, _ := serverPubKey.Curve.ScalarMult(serverPubKey.X, serverPubKey.Y, clientPrivKey.D.Bytes())
	sharedKeyBytes := sharedX.Bytes()
	// Pad to 32 bytes if needed.
	if len(sharedKeyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(sharedKeyBytes):], sharedKeyBytes)
		sharedKeyBytes = padded
	}
	return hex.EncodeToString(sharedKeyBytes), nil
}

// ValidateSharedKey compares the hash of the locally computed shared key with the hashed signature from the server.
func ValidateSharedKey(localSharedKeyHex, serverHashedKeyHex string) bool {
	hasher := sha3.New256()
	// Hash the raw shared key bytes.
	localSharedKeyBytes, err := hex.DecodeString(localSharedKeyHex)
	if err != nil {
		return false
	}
	hasher.Write(localSharedKeyBytes)
	localHash := hex.EncodeToString(hasher.Sum(nil))
	return localHash == serverHashedKeyHex
}

// DecryptChallenge decrypts the challenge using the shared key and provided nonce.
func DecryptChallenge(sharedKeyHex, nonceHex, cipherTextHex string) (string, error) {
	// Decode the shared key.
	sharedKeyBytes, err := hex.DecodeString(sharedKeyHex)
	if err != nil {
		return "", err
	}

	// Create AES cipher block.
	block, err := aes.NewCipher(sharedKeyBytes)
	if err != nil {
		return "", err
	}

	// For this example, assume the nonce is 32 bytes as set on the server.
	aesgcm, err := cipher.NewGCMWithNonceSize(block, 32)
	if err != nil {
		return "", err
	}

	// Decode nonce and ciphertext.
	nonceBytes, err := hex.DecodeString(nonceHex)
	if err != nil {
		return "", err
	}
	cipherTextBytes, err := hex.DecodeString(cipherTextHex)
	if err != nil {
		return "", err
	}

	// Decrypt the challenge.
	plainTextBytes, err := aesgcm.Open(nil, nonceBytes, cipherTextBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plainTextBytes), nil
}

// SignChallenge signs the decrypted challenge using the client's private key.
func SignChallenge(challenge string, clientPrivKey *ecdsa.PrivateKey) (string, error) {
	// Hash the challenge (using SHA3-256, as in your verification).
	hasher := sha3.New256()
	hasher.Write([]byte(challenge))
	hashBytes := hasher.Sum(nil)

	// Sign the hash.
	r, s, err := ecdsa.Sign(rand.Reader, clientPrivKey, hashBytes)
	if err != nil {
		return "", err
	}
	// Concatenate r and s (each padded to 32 bytes) into a single 64-byte signature.
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	if len(rBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(rBytes):], rBytes)
		rBytes = padded
	}
	if len(sBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(sBytes):], sBytes)
		sBytes = padded
	}
	sigBytes := append(rBytes, sBytes...)
	return hex.EncodeToString(sigBytes), nil
}
