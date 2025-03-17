package user

import (
	"context"
	"sync"

	"keyless-auth/services"
)

type repository struct {
	db *services.RedisClient
	m  sync.Mutex
}

func NewUser(db *services.RedisClient) Repo {
	return &repository{
		db: db,
		m:  sync.Mutex{},
	}
}

type Repo interface {
	GetUserByID(id string) (bool, error)

	SavePubKeyUser(ctx context.Context, user *UserWithPubKey) error
	GetPubKeyUserWithID(id string) (*UserWithPubKey, error)

	SaveoAuthUser(user *OAuthUser) error
	GetoAuthUser(ctx context.Context, id string) (*OAuthUser, error)
}

type Cached interface {
	Getter()
	Setter()
}

func (r *repository) Getter() {
	return
}

func (r *repository) Setter() {}

func (r *repository) GetUserByID(id string) (bool, error) {
	return r.db.Client.SIsMember(context.Background(), "user:"+id, id).Result()
}
