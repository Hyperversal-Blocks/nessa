package repository

import (
	"sync"

	"keyless-auth/services"
)

type walletRepo struct {
	db *services.RedisClient
	m  sync.Mutex
}

func NewWalletRepository(db *services.RedisClient) WalletRepository {
	return &walletRepo{
		db: db,
		m:  sync.Mutex{},
	}
}

type WalletRepository interface {
}
