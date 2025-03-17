package signer

import (
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"

	"keyless-auth/signer/signerMock"
)

func DefaultMockOptions(t *testing.T) []signerMock.Option {
	t.Helper()

	key, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)

	return []signerMock.Option{
		signerMock.WithGetPrivateKeyFunc(func() *ecdsa.PrivateKey {
			return key
		}),
		signerMock.WithGetPublicKeyFunc(func() *ecdsa.PublicKey {
			return &key.PublicKey
		}),
		signerMock.WithGetSharedKeyFunc(func(their ecdsa.PublicKey) [32]byte {
			sx, sy := their.ScalarMult(their.X, their.Y, key.D.Bytes())

			hashed := crypto.Keccak256(sx.Bytes(), sy.Bytes())

			var sharedK [32]byte
			copy(sharedK[:], hashed)
			return sharedK
		}),
		signerMock.WithGenNonceFunc(func() string {
			nonce := make([]byte, 12)
			rand.Read(nonce)
			return hex.EncodeToString(nonce)
		}),
		signerMock.WithEncryptAndGetChallengeHash(func(k string, message string) (string, string, error) {
			return k, "", nil
		}),

		signerMock.WithDecryptMessageFunc(func(sharedK string, cipherText string) (string, error) {
			deciphered := "this is a message"
			return deciphered, nil
		}),
		signerMock.WithSignFunc(func(hash string) (string, error) {
			return hash, nil
		}),
		signerMock.WithVerifySignatureFunc(func(pub ecdsa.PublicKey, signature, messageHash string) bool {
			return true
		}),
		signerMock.WithBytesFromPublicKeyFunc(func(k *ecdsa.PublicKey) []byte {
			return crypto.FromECDSAPub(k)
		}),
		signerMock.WithPublicKeyFromBytesFunc(func(b []byte) (*ecdsa.PublicKey, error) {
			return crypto.UnmarshalPubkey(b)
		}),
		signerMock.WithCipherModeFunc(func(key string) (cipher.AEAD, error) {
			return nil, nil
		}),
	}
}

func TestCrypto(t *testing.T) {
	t.Parallel()

	yourSigner := signerMock.New(DefaultMockOptions(t)...)
	theirSigner := signerMock.New(DefaultMockOptions(t)...)

	msg := "this is a message"
	msgToByte := []byte(msg)

	t.Run("gets shared key", func(t *testing.T) {
		t.Parallel()

		sharedKey := yourSigner.GetSharedKey(*theirSigner.GetPublicKey())
		if sharedKey[:] == nil {
			t.Fatal("expected shared key, got\nsharedKey: ", sharedKey)
		}
	})

	t.Run("gets nonce", func(t *testing.T) {
		t.Parallel()

		nonce := yourSigner.GenNonce()
		if len(nonce) == 0 {
			t.Fatal("expected nonce, got\nnonce: ", nonce)
		}
	})

	t.Run("encrypt-decrypt", func(t *testing.T) {
		t.Parallel()

		sharedKey := yourSigner.GetSharedKey(*theirSigner.GetPublicKey())
		if sharedKey[:] == nil {
			t.Fatal("expected shared key, got\nsharedKey: ", sharedKey)
		}

		nonce := yourSigner.GenNonce()
		if len(nonce) == 0 {
			t.Fatal("expected nonce, got\nnonce: ", nonce)
		}

		ciphered, hashed, err := yourSigner.EncryptAndGetChallengeHash(hex.EncodeToString(sharedKey[:]), nonce, hex.EncodeToString(msgToByte))
		if err != nil || len(ciphered) == 0 || len(hashed) == 0 {
			t.Fatal("expected hashed, got: ", hashed,
				"\nexpected ciphered, got: ", ciphered,
				"\nexpected error to be nil, got: ", err)
		}

		theirSharedKey := theirSigner.GetSharedKey(*yourSigner.GetPublicKey())

		deciphered, err := theirSigner.DecryptMessage(hex.EncodeToString(theirSharedKey[:]), ciphered)
		if err != nil {
			t.Fatal("unexpected err: ", err)
		}

		if deciphered != msg {
			t.Fatal("expected messages to be same",
				"\noriginal: ", msg,
				"\ndeciphered: ", deciphered)
		}

		signature, err := yourSigner.Sign(hashed)
		if err != nil {
			t.Fatal("unexpected err: ", err)
		}

		isValid := theirSigner.VerifySignature(*yourSigner.GetPublicKey(), signature, hashed[:])
		if !isValid {
			t.Fatal("expected valid")
		}
	})

	t.Run("verify shared keys matching", func(t *testing.T) {
		t.Parallel()

		theirSharedKey := theirSigner.GetSharedKey(*yourSigner.GetPublicKey())
		yourSharedKey := yourSigner.GetSharedKey(*theirSigner.GetPublicKey())

		fmt.Println(theirSharedKey)
		fmt.Println(yourSharedKey)

		if theirSharedKey != yourSharedKey {
			t.Fatal("expected both to be same\n",
				"their secret: ", hex.EncodeToString(theirSharedKey[:]),
				"\nyour secret: ", hex.EncodeToString(yourSharedKey[:]))
		}
	})

	t.Run("get public key from bytes", func(t *testing.T) {
		t.Parallel()

		publicKeyInBytes := yourSigner.BytesFromPublicKey(yourSigner.GetPublicKey())
		publicKeyString := hex.EncodeToString(publicKeyInBytes)

		pubKeyBytes, _ := hex.DecodeString(publicKeyString)
		pubKey, _ := yourSigner.PublicKeyFromBytes(pubKeyBytes)
		if !yourSigner.GetPublicKey().Equal(pubKey) {
			t.Fatal("original Key:\n", yourSigner.GetPublicKey(), " \nGot:\n", pubKey)
		}

		fmt.Println(pubKey)
		fmt.Println(hex.EncodeToString(publicKeyInBytes))
	})
}

// TestEncryptAndDecryptChallenge verifies that EncryptAndGetChallengeHash
// produces a ciphertext from which the original message can be recovered.
func TestEncryptAndDecryptChallenge(t *testing.T) {
	// Create a new signer instance with a generated private key.
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
	s := &signer{PrivateKey: privKey}

	pubkeyStr := hex.EncodeToString(crypto.FromECDSAPub(&privKey.PublicKey))
	privKeyStr := hex.EncodeToString(crypto.FromECDSA(privKey))

	fmt.Println("pubKeyStr:", pubkeyStr)
	fmt.Println("privKeyStr:", privKeyStr)
	// Create a random shared key (simulate the ECDH-derived shared key).
	sharedKeyBytes := make([]byte, 32)
	if _, err := rand.Read(sharedKeyBytes); err != nil {
		t.Fatalf("failed to generate shared key: %v", err)
	}
	sharedKey := hex.EncodeToString(sharedKeyBytes)

	// Define the test message (challenge).
	testMessage := "This is a test challenge"

	// Encrypt the test message.
	challengeHash, combined, err := s.EncryptAndGetChallengeHash(sharedKey, testMessage)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Decode the combined hex string to raw bytes.
	combinedBytes, err := hex.DecodeString(combined)
	if err != nil {
		t.Fatalf("failed to decode combined ciphertext: %v", err)
	}

	// Get the AES-GCM instance (to determine the nonce size).
	aesgcm, err := s.GetCipherMode(sharedKey)
	if err != nil {
		t.Fatalf("failed to get cipher mode: %v", err)
	}
	nonceSize := aesgcm.NonceSize()
	if len(combinedBytes) < nonceSize {
		t.Fatalf("combined data length (%d) is less than nonce size (%d)", len(combinedBytes), nonceSize)
	}

	// Extract the nonce (first nonceSize bytes) and ciphertext (the remainder).
	nonce := combinedBytes[:nonceSize]
	ciphertext := combinedBytes[nonceSize:]

	// Decrypt the ciphertext using the extracted nonce.
	decrypted, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		t.Fatalf("failed to decrypt ciphertext: %v", err)
	}
	if string(decrypted) != testMessage {
		t.Fatalf("decrypted message mismatch; got '%s', want '%s'", string(decrypted), testMessage)
	}

	// Verify that the challenge hash equals SHA3-256(sharedKey).
	hasher := sha3.New256()
	_, err = hasher.Write([]byte(sharedKey))
	if err != nil {
		t.Fatalf("failed to hash shared key: %v", err)
	}
	expectedHash := hex.EncodeToString(hasher.Sum(nil))
	if challengeHash != expectedHash {
		t.Fatalf("challenge hash mismatch; got '%s', want '%s'", challengeHash, expectedHash)
	}
}

func TestWithPostman(t *testing.T) {
	t.Parallel()
	t.Run("postman", func(t *testing.T) {
		t.Parallel()

		theirPubKey := "044de4a1298d7dc09695a3f50ac75bc47963297a79ed72e518e9b51e2b14515e042b440168c8fff10a7ecccc446a59cf528825c9b8039ce9ac09c4e5eadaffea71"
		_ = "04cc92aea26ad08c582e3d0b57d2d652646594472932de6b4b8828f6a0d5dde3f10f9accbf31303527d2caec36c1a910fe9ccd3aac1c7b55ad816f88f41ac0d6e2"
		_ = "490c6f6ac8c68bca0f77134854fc06067badf0ed6907d11b459218a24f407b2f6d3affc7f606420e2891162ce61fd9ab5282b5bc950018495a4d0cef5f277cde"
		_ = "2dc0a498f14122192fb487f6c6f88ef0c09f5670e0fcd78b225036bfef33e2ca"
		challenge := "c2e2176f7898c064d217c21326410548d8ed9e50ea837bab97038eb39b9a55cae9bc01dc718931f433ef703763c9fede1fd058483c026316010a711937d65fa9"
		privKey := "3c31f2be3b9cd58421e99997f52098374a2ca1162185dac2de520eeda53bdeae"

		s, _ := New()

		_theirPubKey, _ := s.PublicKeyFromBytes(theirPubKey)
		_privKey, _ := crypto.HexToECDSA(privKey)
		_ = _privKey.Public()

		// if !s.VerifySignature(*_theirPubKey, signature, hashedSignature) {
		// 	t.Fatal("failed to verify signature")
		// }

		sharedKey, _ := _theirPubKey.Curve.ScalarMult(_theirPubKey.X, _theirPubKey.Y, _privKey.D.Bytes())
		_sharedKey := hex.EncodeToString(sharedKey.Bytes())
		decryptedMsg, _ := s.DecryptMessage(_sharedKey, challenge)

		fmt.Println(decryptedMsg)
	})
}
