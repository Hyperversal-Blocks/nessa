package session

import (
	"context"
	"time"
)

type Store interface {
	UserExistsForSession(ctx context.Context, userPubKey string) (bool, error)
	EphemeralKeyExists(ctx context.Context, userPubKey, ephemeralPubKey string) (bool, error)
	SaveChallenge(ctx context.Context, record *ChallengeRecord) error
	GetLatestChallenge(ctx context.Context, userPubKey string) (string, error)
	MarkUserVerified(ctx context.Context, userPubKey string) error
	IsUserVerified(ctx context.Context, userPubKey string) (bool, error)
	GetActiveChallengeForUser(ctx context.Context, userPubKey string) (*ChallengeRecord, error)
	DiscardExistingChallenge(
		ctx context.Context,
		userPubKey string) error
	ExtendTTL(ctx context.Context, keys []string, extension time.Duration) error
}
