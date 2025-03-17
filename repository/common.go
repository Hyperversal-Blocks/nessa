package repository

import (
	"fmt"
	"time"

	"github.com/wealdtech/go-merkletree"
)

// MerkleNode describes a single node in the Merkle tree.
type MerkleNode struct {
	ID          string // e.g. a UUID
	Hash        string // Hash of this leaf/node
	Position    uint64
	ProofHashes []string
	TreeRoot    string // Merkle root after insertion
	PrevRoot    string // Optional: the previous root
	CreatedAt   time.Time
	UID         string // which user this node belongs to
}

// -----------------TODO

// GlobalMerkleObject is for future reference
type GlobalMerkleObject struct {
	Node *MerkleNode            `json:"node"`
	Tree *merkletree.MerkleTree `json:"tree"`
}

func (o *GlobalMerkleObject) ToChildren() (*MerkleNode, *merkletree.MerkleTree, error) {
	return nil, nil, nil
}

func (o *GlobalMerkleObject) ToParent(*MerkleNode, *merkletree.MerkleTree, error) error {
	return nil
}

type Object struct {
	Tree      *Tree     `json:"tree,inline"`
	User      *User     `json:"user,inline"`
	CreatedAt time.Time `json:"created_at"`
}

type Tree struct {
	Leaf          string            `json:"leaf"` // leaf is the credential
	Index         uint64            `json:"index"`
	ProofElements []string          `json:"proof_elements"`
	Root          string            `json:"root"`      // Merkle root after insertion
	PrevRoot      string            `json:"prev_root"` // Optional: the previous root
	Proof         *merkletree.Proof `json:"proof"`
}

type User struct {
	ID           string `json:"id"`
	Wallet       string `json:"wallet"`
	CredentialID string `json:"credential_id"`
}

func (o *Object) Key() string {
	return fmt.Sprintf("uid:root:credential")
}
