package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/wealdtech/go-merkletree"

	"keyless-auth/repository"
	"keyless-auth/repository/user"
	"keyless-auth/service"
)

type GenerateTreeRequest struct {
	HashedCredential string `json:"hashed_credential"`
}

type GenerateTreeResponse struct {
	Object repository.Object `json:"object,inline"`
}

type CredentialRequest struct {
	HashedCredential string `json:"hashed_credential"`
	UserID           string `json:"user_id"`
}

type CredentialResponse struct {
	MerkleRoot    string            `json:"merkle_root"`
	WalletAddress string            `json:"wallet_address"`
	Proof         *merkletree.Proof `json:"proof"`
	Leaf          string            `json:"leaf"`
}

type MerkleRootResponse struct {
	MerkleRoot string `json:"merkle_root"`
	NumLeaves  int    `json:"num_leaves"`
}

type MerkleProofResponse struct {
	Proof *merkletree.Proof `json:"proof"`
}

type CredentialsHandler struct {
	credRepo   *repository.CredentialsRepository
	merkleTree *service.MerkleTreeService
	userRepo   user.Repo
	walletRepo repository.WalletRepository
}

func NewCredentialsHandler(credRepo *repository.CredentialsRepository, walletRepo repository.WalletRepository) *CredentialsHandler {
	return &CredentialsHandler{
		credRepo:   credRepo,
		walletRepo: walletRepo,
		merkleTree: service.NewMerkleTreeService(credRepo),
	}
}

func (h *CredentialsHandler) GenerateMerkleProof(w http.ResponseWriter, r *http.Request) {
	credential := mux.Vars(r)["credential"]
	// TODO: with existing credential
	treeObj, err := h.merkleTree.GenerateTree(credential)
	if err != nil {
		http.Error(w, "failed to generate merkle proof", http.StatusInternalServerError)
		return
	}

	err = h.credRepo.SaveCredentialAndNode(context.Background(), credential, treeObj.Root, &repository.MerkleNode{
		ID:          uuid.New().String(),
		Hash:        treeObj.Leaf,
		Position:    treeObj.Index,
		ProofHashes: treeObj.ProofElements,
		TreeRoot:    treeObj.Root,
		PrevRoot:    treeObj.PrevRoot,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		log.Println("failed to storage credential and node: ", err)
		http.Error(w, "failed to storage credential and node", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(MerkleProofResponse{
		Proof: treeObj.Proof,
	})
}

func GenerateWalletAddress() (string, []byte, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return "", nil, err
	}

	return crypto.PubkeyToAddress(privKey.PublicKey).Hex(), privKey.D.Bytes(), nil
}

func (h *CredentialsHandler) GenerateCredential(w http.ResponseWriter, r *http.Request) {
	var req CredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// check if user exists
	exists, err := h.userRepo.GetUserByID(req.UserID)
	if err != nil {
		http.Error(w, "unable to process user", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// check if credential already exists
	if exists, err := h.credRepo.Exists(req.HashedCredential); err != nil || exists {
		// TODO: better return 200 with message to fetch the wallet address or credential on a different endpoint
		http.Error(w, "Credential already exists", http.StatusBadRequest)
		return
	}

	treeObj, err := h.merkleTree.GenerateTree(req.HashedCredential)
	if err != nil {
		http.Error(w, "Failed to generate merkle tree root", http.StatusInternalServerError)
		return
	}

	nodeObj := &repository.MerkleNode{
		ID:          uuid.New().String(),
		Hash:        treeObj.Leaf,
		Position:    treeObj.Index,
		ProofHashes: treeObj.ProofElements,
		TreeRoot:    treeObj.Root,
		PrevRoot:    treeObj.PrevRoot,
		CreatedAt:   time.Now(),
	}

	// "SaveCredentialAndNode" if you want to storage the root. We can also storage node only.
	err = h.credRepo.SaveCredentialAndNode(context.Background(), treeObj.Leaf, treeObj.Root, nodeObj)
	if err != nil {
		log.Println("failed to storage credential and node: ", err)
		http.Error(w, "failed to storage credential and node", http.StatusInternalServerError)
		return
	}

	err = h.credRepo.SetMostRecentMerkleNode(context.Background(), nodeObj)
	if err != nil {
		log.Println("failed to storage recent node: ", err)
		http.Error(w, "failed to storage recent node", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(CredentialResponse{
		MerkleRoot:    treeObj.Root,
		WalletAddress: "",
		Proof:         treeObj.Proof,
		Leaf:          treeObj.Leaf,
	})
}

func (h *CredentialsHandler) GetMerkleRoot(w http.ResponseWriter, r *http.Request) {
	node, err := h.credRepo.GetMostRecentMerkleNode(context.Background())
	if err != nil {
		log.Println("failed to fetch recent node: ", err)
		http.Error(w, "failed to fetch recent node", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(MerkleRootResponse{
		MerkleRoot: node.TreeRoot,
		NumLeaves:  len(node.ProofHashes),
	})
}

func (h *CredentialsHandler) GenerateTreeObject(w http.ResponseWriter, r *http.Request) {
	var req GenerateTreeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.HashedCredential == "" {
		http.Error(w, "invalid or empty request object", http.StatusBadRequest)
		return
	}

	// check if credential already exists
	if exists, err := h.credRepo.Exists(req.HashedCredential); err != nil || exists {
		http.Error(w, "Credential already exists", http.StatusUnprocessableEntity)
		return
	}

	// TODO: get user from session state
	// TODO: get user wallet

	treeObj, err := h.merkleTree.GenerateTree(req.HashedCredential)
	if err != nil {
		http.Error(w, "Failed to generate merkle tree root", http.StatusInternalServerError)
		return
	}

	object := &repository.Object{
		Tree:      treeObj,
		User:      nil,
		CreatedAt: time.Now().UTC(),
	}

	err = h.credRepo.AddToTree(context.Background(), object)
	if err != nil {
		log.Println("failed to storage credential and node: ", err)
		http.Error(w, "failed to storage credential and node", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&object)
}

// Register with a credential
// We generate a wallet
// We generate merkle tree -> proof
// merkle tree/proof
