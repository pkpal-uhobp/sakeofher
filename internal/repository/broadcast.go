package repository

import "sakeofher/internal/repository/tx"

type BroadcastRepository struct{ tx *tx.Manager }

func NewBroadcastRepository(txManager *tx.Manager) *BroadcastRepository {
	return &BroadcastRepository{tx: txManager}
}
