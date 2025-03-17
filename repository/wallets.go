package repository

import (
	"sync"

	"github.com/hblocks/nessa/services"
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
