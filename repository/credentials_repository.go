package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"keyless-auth/services"
)

// CredentialsRepository manages credential<->wallet<->user data.
type CredentialsRepository struct {
	db *services.RedisClient
}

func NewCredentialsRepository(db *services.RedisClient) *CredentialsRepository {
	return &CredentialsRepository{db: db}
}

// Exists checks if a credential is in the global set of credentials.
func (r *CredentialsRepository) Exists(credential string) (bool, error) {
	ctx := context.Background()
	return r.db.Client.SIsMember(ctx, "merkle:credentials:set", credential).Result()
}

// SaveCredentialAndNode is a high-level method that:
//   - Adds the credential to a global set (for existence checks).
//   - Appends the MerkleNode to a list keyed by credentialID.
//   - Appends the Merkle root to a *list* of historical roots for that credential.
//
// If you only need the *unique* set of roots, switch to SAdd. But for chronological order, use RPUSH.
func (r *CredentialsRepository) SaveCredentialAndNode(
	ctx context.Context,
	credential string,
	root string,
	node *MerkleNode,
) error {
	if err := r.db.Client.SAdd(ctx, "merkle:credentials:set", credential).Err(); err != nil {
		log.Printf("Failed to add credential %q to global set: %v", credential, err)
		return err
	}

	jsonNode, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal node to JSON: %w", err)
	}

	nodesKey := fmt.Sprintf("merkle:credential:%s:nodes", credential)
	if err := r.db.Client.RPush(ctx, nodesKey, jsonNode).Err(); err != nil {
		log.Printf("Failed to add node JSON to Redis list %q: %v", nodesKey, err)
		return err
	}

	rootsKey := fmt.Sprintf("merkle:credential:%s:roots", credential)
	if err := r.db.Client.RPush(ctx, rootsKey, root).Err(); err != nil {
		log.Printf("Failed to add root %q to Redis list %q: %v", root, rootsKey, err)
		return err
	}

	return nil
}

func (r *CredentialsRepository) SaveMerkleNode(
	ctx context.Context,
	credentialID string,
	node *MerkleNode,
) error {
	nodeJSON, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal MerkleNode: %w", err)
	}

	nodesKey := fmt.Sprintf("merkle:credential:%s:nodes", credentialID)
	if err := r.db.Client.RPush(ctx, nodesKey, nodeJSON).Err(); err != nil {
		return fmt.Errorf("failed to storage MerkleNode in Redis: %w", err)
	}

	return nil
}

// GetNodesByCredential returns all MerkleNodes (in insertion order) for a credential.
func (r *CredentialsRepository) GetNodesByCredential(
	ctx context.Context,
	credential string,
) ([]MerkleNode, error) {
	nodesKey := fmt.Sprintf("merkle:credential:%s:nodes", credential)

	items, err := r.db.Client.LRange(ctx, nodesKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nodes for credential=%q: %w", credential, err)
	}

	var nodes []MerkleNode
	for _, jsonStr := range items {
		var node MerkleNode
		if err := json.Unmarshal([]byte(jsonStr), &node); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node JSON: %w", err)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetLatestMerkleNode returns only the *most recent* node for a credential.
func (r *CredentialsRepository) GetLatestMerkleNode(
	ctx context.Context,
	credential string,
) (*MerkleNode, error) {
	nodesKey := fmt.Sprintf("merkle:credential:%s:nodes", credential)

	items, err := r.db.Client.LRange(ctx, nodesKey, -1, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest node for credential=%q: %w", credential, err)
	}
	if len(items) == 0 {
		return nil, nil // no node found
	}

	var node MerkleNode
	if err := json.Unmarshal([]byte(items[0]), &node); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MerkleNode: %w", err)
	}

	return &node, nil
}

// GetRootsByCredential returns a *chronological list* of roots for that credential.
// RPush --> index 0 is the initial root, last index is the newest root.
func (r *CredentialsRepository) GetRootsByCredential(
	ctx context.Context,
	credential string,
) ([]string, error) {
	rootsKey := fmt.Sprintf("merkle:credential:%s:roots", credential)
	roots, err := r.db.Client.LRange(ctx, rootsKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch root hashes for credential %q: %w", credential, err)
	}
	return roots, nil
}

// GetUserCredentials returns an ordered list of all credentials a user has.
func (r *CredentialsRepository) GetUserCredentials(
	ctx context.Context,
	userID string,
) ([]string, error) {
	userCredsKey := fmt.Sprintf("merkle:user:%s:credentials", userID)
	creds, err := r.db.Client.LRange(ctx, userCredsKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credentials for user %q: %w", userID, err)
	}
	return creds, nil
}

func (r *CredentialsRepository) AddGlobalCredential(
	ctx context.Context,
	credential string,
) error {
	// We do RPUSH so the oldest is at index 0, newest at the last index
	return r.db.Client.RPush(ctx, "merkle:global:credentials", credential).Err()
}

func (r *CredentialsRepository) GetAllGlobalCredentials(
	ctx context.Context,
) ([]string, error) {
	// Return them from index 0..-1, oldest to newest
	return r.db.Client.LRange(ctx, "merkle:global:credentials", 0, -1).Result()
}

func (r *CredentialsRepository) GetMostRecentMerkleNode(ctx context.Context) (*MerkleNode, error) {
	key := "merkle:global:nodes"

	result, err := r.db.Client.LIndex(ctx, key, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the most recent node: %w", err)
	}
	var node MerkleNode
	if err := json.Unmarshal([]byte(result), &node); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the most recent node: %w", err)
	}
	return &node, nil
}

func (r *CredentialsRepository) SetMostRecentMerkleNode(ctx context.Context, node *MerkleNode) error {
	nodeJSON, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal MerkleNode: %w", err)
	}

	key := "merkle:global:nodes"
	if err := r.db.Client.RPush(ctx, key, nodeJSON).Err(); err != nil {
		return fmt.Errorf("failed to storage MerkleNode in Redis: %w", err)
	}

	return nil
}

// --------------------- TODO---------------------------------------------

func (r *CredentialsRepository) SetGlobalMerkleObject(
	ctx context.Context,
	obj *GlobalMerkleObject,
) error {
	key := "merkle:global"
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return r.db.Client.RPush(ctx, key, data).Err()
}

func (r *CredentialsRepository) GetGlobalMerkleObject() (*GlobalMerkleObject, error) {
	// Make sure this key matches the one you used with RPUSH
	key := "merkle:global"

	// LRange(key, -1, -1) returns only the last element in the list
	obj, err := r.db.Client.LRange(context.Background(), key, -1, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the most recent node: %w", err)
	}

	// Check if there's at least one element
	if len(obj) == 0 {
		return nil, errors.New("no nodes found")
	}

	// Unmarshal JSON into your MerkleNode struct
	var _obj GlobalMerkleObject
	if err := json.Unmarshal([]byte(obj[0]), &_obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the most recent node: %w", err)
	}

	return &_obj, nil
}

// AddSingleCredentialToWallet adds *one* credential to a wallet’s list of credentials.
// It also ensures that the credential->wallet mapping is set in a Redis hash.
func (r *CredentialsRepository) AddSingleCredentialToWallet(
	ctx context.Context,
	wallet string,
	credential string,
) error {
	mappedWallet, err := r.db.Client.HGet(ctx, "merkle:credentialToWallet", credential).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get credential owner for %q: %w", credential, err)
	}

	if mappedWallet != "" && mappedWallet != wallet {
		return fmt.Errorf("credential %q is already owned by wallet %q", credential, mappedWallet)
	}

	if err := r.db.Client.HSet(ctx, "merkle:credentialToWallet", credential, wallet).Err(); err != nil {
		return fmt.Errorf("failed to set credential->wallet mapping: %w", err)
	}

	walletKey := fmt.Sprintf("merkle:wallet:%s:credentials", wallet)
	if err := r.db.Client.RPush(ctx, walletKey, credential).Err(); err != nil {
		return fmt.Errorf("failed to add credential %q to wallet %q: %w", credential, wallet, err)
	}

	return nil
}

// SetCredentialsForWallet is the "setter" method to overwrite a wallet’s credential list
// with a new collection, in order. (Example usage: if you want to storage multiple at once.)
// This is for future reference.
func (r *CredentialsRepository) SetCredentialsForWallet(
	ctx context.Context,
	wallet string,
	credentials []string,
) error {
	// Key for wallet’s credential list
	walletKey := fmt.Sprintf("merkle:wallet:%s:credentials", wallet)

	// 1) Delete the existing list entirely
	if err := r.db.Client.Del(ctx, walletKey).Err(); err != nil {
		return fmt.Errorf("failed to clear existing credentials for wallet %q: %w", wallet, err)
	}

	// 2) RPush each credential in the order provided
	for _, c := range credentials {
		if err := r.db.Client.RPush(ctx, walletKey, c).Err(); err != nil {
			return fmt.Errorf("failed to add credential %q to wallet %q: %w", c, wallet, err)
		}
		// Also optionally set the hash mapping if needed:
		// r.db.Client.HSet(ctx, "merkle:credentialToWallet", c, wallet)
	}

	return nil
}

// GetCredentialsForWallet returns the *ordered* list of credentials for a wallet.
func (r *CredentialsRepository) GetCredentialsForWallet(
	ctx context.Context,
	wallet string,
) ([]string, error) {
	walletKey := fmt.Sprintf("merkle:wallet:%s:credentials", wallet)
	creds, err := r.db.Client.LRange(ctx, walletKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credentials for wallet %q: %w", wallet, err)
	}
	return creds, nil
}

// AddCredentialToUser ensures a credential is owned by the user’s wallet and maintains order.
func (r *CredentialsRepository) AddCredentialToUser(
	ctx context.Context,
	userID string,
	wallet string,
	credentialID string,
) error {
	currentWallet, err := r.db.Client.HGet(ctx, "merkle:credentialToWallet", credentialID).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to fetch wallet for credential %q: %w", credentialID, err)
	}
	if currentWallet != "" && currentWallet != wallet {
		return fmt.Errorf("credential %q is already bound to different wallet %q", credentialID, currentWallet)
	}

	if err := r.db.Client.HSet(ctx, "merkle:credentialToWallet", credentialID, wallet).Err(); err != nil {
		return fmt.Errorf("failed to map credential->wallet in Redis: %w", err)
	}

	userCredsKey := fmt.Sprintf("merkle:user:%s:credentials", userID)
	if err := r.db.Client.RPush(ctx, userCredsKey, credentialID).Err(); err != nil {
		return fmt.Errorf("failed to add credential %q to user %q: %w", credentialID, userID, err)
	}

	walletKey := fmt.Sprintf("merkle:wallet:%s:credentials", wallet)
	if err := r.db.Client.RPush(ctx, walletKey, credentialID).Err(); err != nil {
		return fmt.Errorf("failed to add credential %q to wallet %q: %w", credentialID, wallet, err)
	}

	return nil
}

// Test
func (r *CredentialsRepository) AddToTree(
	ctx context.Context,
	obj *Object,
) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return r.db.Client.RPush(ctx, obj.Key(), data).Err()
}

func (r *CredentialsRepository) GetFromTree(
	ctx context.Context,
) ([]Object, error) {

	var _obj Object
	// Return them from index 0..-1, oldest to newest
	obj, err := r.db.Client.LRange(ctx, _obj.Key(), -1, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch objects for key: %s : %w", _obj.Key(), err)
	}

	var nodes []Object
	for _, jsonStr := range obj {
		var node Object
		if err := json.Unmarshal([]byte(jsonStr), &node); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node JSON: %w", err)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (r *CredentialsRepository) RecentMerkleNode(ctx context.Context) (*Object, error) {
	obj := Object{}

	key := obj.Key()

	result, err := r.db.Client.LIndex(ctx, key, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the most recent node: %w", err)
	}
	var node Object
	if err := json.Unmarshal([]byte(result), &node); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the most recent node: %w", err)
	}
	return &node, nil
}
