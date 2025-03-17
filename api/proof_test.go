package api

import (
	"crypto/md5"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	witness2 "github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"

	"keyless-auth/circuit"
)

func TestCompile(t *testing.T) {

	t.Run("compile-success", func(t *testing.T) {
		assignment := circuit.ZKAuthCircuit{
			Leaf:          md5.Sum([]byte("aae457593db8c6ab406c81939d4ffb39e9ac16aeceeb6e109d1593aff3b91ecd")),
			Root:          md5.Sum([]byte(("ed6cfd06d0c37b1f964a2e63e5dce2bab107297c4a5518392d08a2cea24794dc"))),
			ProofElements: []frontend.Variable{md5.Sum([]byte("")), md5.Sum([]byte("72340a9e5f50b5ec542c0f1fd30c1bc47a1f975faabcd3d7748973e8c5fd75d4"))},
			ProofIndex:    1,
		}

		r1cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &assignment)
		if err != nil {
			t.Error("Failed to compile")
		}

		pk, _, err := groth16.Setup(r1cs)
		if err != nil {
			t.Error(
				"Failed to setup",
				err)
		}
		w, err := witness2.New(ecc.BN254.ScalarField())

		prf, err := groth16.Prove(r1cs, pk, w)
		if err != nil {
			t.Error(err)
		}
		t.Log(prf)
	})
}
