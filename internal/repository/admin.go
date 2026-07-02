package repository

import "sakeofher/internal/repository/tx"

type AdminRepository struct{ tx *tx.Manager }

func NewAdminRepository(txManager *tx.Manager) *AdminRepository {
	return &AdminRepository{tx: txManager}
}
