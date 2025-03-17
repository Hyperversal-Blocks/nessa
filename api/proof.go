package api

import (
	"encoding/json"
	"net/http"

	"keyless-auth/circuit"
	"keyless-auth/domain"
	"keyless-auth/repository"

	"github.com/consensys/gnark/backend/groth16"
)

type ProofRequest struct {
	Leaf     string   `json:"leaf"`     // Leaf hash
	Root     string   `json:"root"`     // Merkle root
	Siblings []string `json:"siblings"` // Sibling hashes
	Position int      `json:"position"` // Position (0 = left, 1 = right)
}

type ProofResponse struct {
	Proof *groth16.Proof `json:"proof"`
}

type ProofHandler struct {
	walletRepo repository.WalletRepository
}

func NewProofHandler(walletRepo repository.WalletRepository) *ProofHandler {
	return &ProofHandler{
		walletRepo: walletRepo,
	}
}

func (h *ProofHandler) GenerateProof(w http.ResponseWriter, r *http.Request) {
	var req ProofRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	proof, err := circuit.CompileCircuit(domain.Proof{
		Leaf:     "0x" + req.Leaf,
		Root:     "0x" + req.Root,
		Siblings: req.Siblings,
		Position: req.Position,
	})
	if err != nil {
		http.Error(w, "Failed to compile circuit", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(ProofResponse{Proof: proof})
}
