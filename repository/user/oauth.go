package user

import (
	"context"
)

type OAuthUser struct {
	ID             string `json:"id"`
	GoogleID       string `json:"google_id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture"`
	AccessToken    string `json:"-"`
	RefreshToken   string `json:"-"`
}

func (r *repository) SaveoAuthUser(user *OAuthUser) error {
	if r.m.TryLock() {
		r.m.Lock()
	}
	defer r.m.Unlock()
	ctx := context.Background()
	return r.db.Client.Set(ctx, "user:"+user.ID, user, 0).Err()
}

func (r *repository) GetoAuthUser(ctx context.Context, id string) (*OAuthUser, error) {
	if r.m.TryLock() {
		r.m.Lock()
	}
	defer r.m.Unlock()

	user := &OAuthUser{}

	if err := r.db.Client.Get(ctx, "user:"+id).Scan(&user); err != nil {
		return nil, err
	}

	return user, nil
}
