package session

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// TODO: mutex and handling concurrent calls.

type ChallengeRecord struct {
	ServerSignature   string
	UserPubKey        string
	EphemeralPubKey   string
	EphemeralPrivKey  string
	HashedSharedKey   string
	CipheredChallenge string
	CreatedAt         int64
	TTL               int64
}

func (c *ChallengeRecord) String() string {
	return fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s|%d|%d",
		c.EphemeralPubKey,
		c.EphemeralPrivKey,
		c.ServerSignature,
		c.UserPubKey,
		c.HashedSharedKey,
		c.CipheredChallenge,
		c.CreatedAt,
		c.TTL,
	)
}

type RedisSessionStore struct {
	rdb *redis.Client
}

func NewRedisSessionStore(rdb *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{rdb: rdb}
}

func (s *RedisSessionStore) UserExistsForSession(ctx context.Context, userPubKey string) (bool, error) {
	key := "UserChallenges:" + userPubKey
	// If the list doesn't exist or is empty, it means no challenge has been saved
	length, err := s.rdb.LLen(ctx, key).Result()
	if errors.Is(err, redis.Nil) || length == 0 {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *RedisSessionStore) EphemeralKeyExists(ctx context.Context, userPubKey, ephemeralPubKey string) (bool, error) {
	setKey := "EphemeralKeys:" + userPubKey
	exists, err := s.rdb.SIsMember(ctx, setKey, ephemeralPubKey).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *RedisSessionStore) SaveChallenge(
	ctx context.Context,
	record *ChallengeRecord,
) error {

	listKey := "UserChallenges:" + record.UserPubKey
	setKey := "EphemeralKeys:" + record.UserPubKey

	state := record.String()

	if err := s.rdb.RPush(ctx, listKey, state).Err(); err != nil {
		return err
	}
	if err := s.rdb.Expire(ctx, listKey, time.Duration(record.TTL)*time.Second).Err(); err != nil {
		return err
	}

	// Also store ephemeralPubKey in the set to quickly check existence
	if err := s.rdb.SAdd(ctx, setKey, record.EphemeralPubKey).Err(); err != nil {
		return err
	}
	if err := s.rdb.Expire(ctx, setKey, time.Duration(record.TTL)*time.Second).Err(); err != nil {
		return err
	}

	return nil
}

func (s *RedisSessionStore) GetLatestChallenge(ctx context.Context, userPubKey string) (string, error) {
	listKey := "UserChallenges:" + userPubKey
	// Get last item (most recent)
	result, err := s.rdb.LRange(ctx, listKey, -1, -1).Result()
	if errors.Is(err, redis.Nil) || len(result) == 0 {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return result[0], nil
}

func (s *RedisSessionStore) MarkUserVerified(ctx context.Context, userPubKey string) error {
	return s.rdb.SAdd(ctx, "VerifiedUsers", userPubKey).Err()
}

func (s *RedisSessionStore) IsUserVerified(ctx context.Context, userPubKey string) (bool, error) {
	res, err := s.rdb.SIsMember(ctx, "VerifiedUsers", userPubKey).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return res, nil
}

func (s *RedisSessionStore) GetActiveChallengeForUser(ctx context.Context, userPubKey string) (*ChallengeRecord, error) {
	listKey := "UserChallenges:" + userPubKey
	// Grab the last item in the list
	result, err := s.rdb.LRange(ctx, listKey, -1, -1).Result()
	if errors.Is(err, redis.Nil) || len(result) == 0 {
		return nil, nil // No active challenge found
	}
	if err != nil {
		return nil, err
	}
	parts := strings.Split(result[0], "|")
	if len(parts) < 8 {
		return nil, errors.New("invalid challenge record format")
	}

	ttl, _ := strconv.ParseInt(parts[7], 10, 64)
	createdAt, _ := strconv.ParseInt(parts[6], 10, 64)

	record := &ChallengeRecord{
		EphemeralPubKey:   parts[0],
		EphemeralPrivKey:  parts[1],
		ServerSignature:   parts[2],
		UserPubKey:        parts[3],
		HashedSharedKey:   parts[4],
		CipheredChallenge: parts[5],
		CreatedAt:         createdAt,
		TTL:               ttl,
	}
	return record, nil
}

func (s *RedisSessionStore) ExtendTTL(ctx context.Context, keys []string, extension time.Duration) error {
	for _, key := range keys {
		currTTL, err := s.rdb.TTL(ctx, key).Result()
		if err != nil {
			// If redis.Nil, the key doesn't exist in Redis
			// Return nil or an error depending on your use-case
			if errors.Is(err, redis.Nil) {
				return nil
			}
			return err
		}

		// If the key doesn't exist, TTL is -2
		if currTTL == -2 {
			// Key is gone, so nothing to extend
			return nil
		}

		// If the key exists but has no TTL set, TTL is -1
		if currTTL == -1 {
			// Option A: set a fresh TTL with the extension
			// Option B: return an error or skip
			return s.rdb.Expire(ctx, key, extension).Err()
		}

		// Otherwise, currTTL >= 0
		newTTL := currTTL + extension
		// Make sure newTTL is nonnegative
		if newTTL < 0 {
			// If extension is negative and makes TTL invalid, handle it:
			// either set it to 0 (expire immediately), or treat as no-op, or return an error
			newTTL = 0
		}

		return s.rdb.Expire(ctx, key, newTTL).Err()
	}
	return nil
}

// DiscardExistingChallenge removes the user's active (most recent) challenge
func (s *RedisSessionStore) DiscardExistingChallenge(
	ctx context.Context,
	userPubKey string) error {
	listKey := "UserChallenges:" + userPubKey
	setKey := "EphemeralKeys:" + userPubKey

	err := s.rdb.Del(ctx, listKey).Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to delete old challenge: %w", err)
	}

	err = s.rdb.Del(ctx, setKey).Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to delete old ephemeral key: %w", err)
	}

	return nil
}
