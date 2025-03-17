package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"

	"keyless-auth/repository"
)

type MerkleTreeService struct {
	credRepo *repository.CredentialsRepository
}

func NewMerkleTreeService(credRepo *repository.CredentialsRepository) *MerkleTreeService {
	return &MerkleTreeService{
		credRepo: credRepo,
	}
}

func (s *MerkleTreeService) GenerateTree(nCredential string) (*repository.Tree, error) {
	err := s.credRepo.AddGlobalCredential(context.Background(), hex.EncodeToString([]byte(nCredential)))
	if err != nil {
		return nil, fmt.Errorf("failed to append to global credentials: %w", err)
	}

	credentials, err := s.credRepo.GetAllGlobalCredentials(context.Background())
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to retrieve global credentials: %w", err)
	}

	var data [][]byte
	for _, credential := range credentials {
		hexCredential, err := hex.DecodeString(credential)
		if err != nil {
			return nil, fmt.Errorf("failed to decode credential: %w", err)
		}

		data = append(data, hexCredential)
	}

	tree, err := merkletree.NewUsing(data, keccak256.New(), []byte{})
	if err != nil {
		return nil, fmt.Errorf("failed to build Merkle tree: %w", err)
	}

	proofIndex := uint64(len(data) - 1)
	proof, err := tree.GenerateProof(data[proofIndex])
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %w", err)
	}

	prevRecentNode, err := s.credRepo.RecentMerkleNode(context.Background())
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to get most recent merkle node: %w", err)
	}

	var proofHashes []string
	for _, proofHash := range proof.Hashes {
		proofHashes = append(proofHashes, hex.EncodeToString(proofHash))
	}

	prevRecentRoot := ""
	if prevRecentNode != nil {
		prevRecentRoot = prevRecentNode.Tree.Root
	}

	return &repository.Tree{
		Leaf:          nCredential,
		Index:         proof.Index,
		ProofElements: proofHashes,
		Root:          hex.EncodeToString(tree.Root()),
		PrevRoot:      prevRecentRoot,
		Proof:         proof,
	}, nil
}
