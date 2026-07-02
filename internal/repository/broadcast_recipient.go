package repository

import "sakeofher/internal/repository/tx"

type BroadcastRecipientRepository struct{ tx *tx.Manager }

func NewBroadcastRecipientRepository(txManager *tx.Manager) *BroadcastRecipientRepository {
	return &BroadcastRecipientRepository{tx: txManager}
}
