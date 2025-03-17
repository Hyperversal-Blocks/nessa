package user

import (
	"context"
	"fmt"
)

type UserWithPubKey struct {
	PubKey              string `json:"pub_key"`
	ID                  string `json:"id"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
	CreatedAt           int64  `json:"created_at"`
}

func (c *UserWithPubKey) String() string {
	return fmt.Sprintf(
		"%s|%s|%s|%d",
		c.ID,
		c.PubKey,
		c.EncryptedPrivateKey,
		c.CreatedAt,
	)
}

func (r *repository) GetPubKeyUserWithID(id string) (*UserWithPubKey, error) {
	user := &UserWithPubKey{}

	if err := r.db.Client.Get(context.Background(), "user:"+id).Scan(&user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *repository) SavePubKeyUser(ctx context.Context, user *UserWithPubKey) error {
	setKeyUID := "UID:" + user.ID
	return r.db.Client.SAdd(ctx, setKeyUID, user.String(), 0).Err()
}
