package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	"golang.org/x/crypto/sha3"
)

// GenerateSHAKE256Hashes creates N random 32-byte buffers,
// derives a 32-byte SHAKE256 digest for each, returning them as hex strings.
func GenerateSHAKE256Hashes(count int) ([]string, error) {
	hashes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		// Create random input data
		data := make([]byte, 32)
		if _, err := rand.Read(data); err != nil {
			return nil, err
		}

		// Create a SHAKE256 hasher
		shake := sha3.NewShake256()

		// Write random data
		if _, err := shake.Write(data); err != nil {
			return nil, err
		}

		// Read 32 bytes from the SHAKE stream
		digest := make([]byte, 32)
		if _, err := shake.Read(digest); err != nil {
			return nil, err
		}

		// Convert to hex
		hashes = append(hashes, hex.EncodeToString(digest))
	}
	return hashes, nil
}

// MarshalResultAsJSON indents the final map to pretty JSON (optional convenience).
func MarshalResultAsJSON(result map[string]interface{}) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
