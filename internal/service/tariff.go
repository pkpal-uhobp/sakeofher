package service

import (
	"context"
	"sakeofher/internal/domain"
	"sakeofher/internal/repository"
)

type TariffService struct{ repo *repository.Repositories }

func NewTariffService(repo *repository.Repositories) *TariffService {
	return &TariffService{repo: repo}
}

func (s *TariffService) ListActive(ctx context.Context) ([]domain.Tariff, error) {
	return s.repo.Tariffs.ListActive(ctx)
}
