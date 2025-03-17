package signer

import (
	"testing"
)

func TestSigner(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		signer, err := New()
		if err != nil {
			t.Fatal("error generating signer: ", err)
		}

		if signer == nil {
			t.Fatal("expected signer to not be nil: ", err)
		}

		ethAddress := signer.EthereumAddress()
		if &ethAddress == nil {
			t.Fatal("expected eth address to not be nil: ", err)
		}
	})

	t.Run("incorrect private key", func(t *testing.T) {
		t.Parallel()

		signer, err := New()
		if err == nil {
			t.Fatal("expected error: ", err)
		}

		if signer != nil {
			t.Fatal("expected signer to not be nil: ", err)
		}
	})

	t.Run("incorrect private key", func(t *testing.T) {
		t.Parallel()

		signer, err := New()
		if err == nil {
			t.Fatal("expected error: ", err)
		}

		if signer != nil {
			t.Fatal("expected signer to not be nil: ", err)
		}
	})
}
