package circuit

import (
	"fmt"
	"math/big"
	"strconv"

	"keyless-auth/domain"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

var (
	MAX_DEPTH = 256
)

func GenerateGroth16(assignment *ZKAuthCircuit) (constraint.ConstraintSystem, *groth16.ProvingKey, *groth16.VerifyingKey, error) {
	r1cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, assignment)
	if err != nil {
		return nil, nil, nil, err
	}

	pk, vk, err := groth16.Setup(r1cs)
	if err != nil {
		return nil, nil, nil, err
	}
	return r1cs, &pk, &vk, nil
}

func helper(val string) (*big.Int, error) {
	bigVal, ok := new(big.Int).SetString(val, 16)
	if !ok {
		fmt.Errorf("failed to parse %s", val)
	}
	return bigVal, nil
}

func CompileCircuit(proof domain.Proof) (*groth16.Proof, error) {
	leaf, err := helper(proof.Leaf)
	root, err := helper(proof.Root)

	s1, s2 := new(big.Int), new(big.Int)

	for _, sibling := range proof.Siblings {
		s1, err = helper("0x" + strconv.Itoa(int(sibling[0])))
		if err != nil {
			return nil, err
		}
		s2, err = helper("0x" + strconv.Itoa(int(sibling[0])))
		if err != nil {
			return nil, err
		}
	}

	assignment := ZKAuthCircuit{
		Leaf:          frontend.Variable(leaf),
		Root:          frontend.Variable(root),
		ProofElements: make([]frontend.Variable, len(proof.Siblings)),
		ProofIndex:    frontend.Variable(proof.Position),
	}

	assignment.ProofElements[0] = s1
	assignment.ProofElements[1] = s2

	r1cs, pk, _, err := GenerateGroth16(&assignment)
	if err != nil {
		return nil, err
	}

	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, err
	}

	prf, err := groth16.Prove(r1cs, *pk, witness)
	if err != nil {
		return nil, err
	}

	return &prf, nil
}
