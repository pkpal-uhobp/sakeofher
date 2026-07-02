package repository

import "sakeofher/internal/repository/tx"

type RemnaSyncRepository struct{ tx *tx.Manager }

func NewRemnaSyncRepository(txManager *tx.Manager) *RemnaSyncRepository {
	return &RemnaSyncRepository{tx: txManager}
}
