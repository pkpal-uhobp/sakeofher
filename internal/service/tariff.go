package service

import (
	"context"
	"sakeofher/internal/domain"
	"sakeofher/internal/repository"
)

type tariffService struct{ repo *repository.Repositories }

func NewTariffService(repo *repository.Repositories) TariffService {
	return &tariffService{repo: repo}
}

func (s *tariffService) ListActive(ctx context.Context) ([]domain.Tariff, error) {
	return s.repo.Tariffs.ListActive(ctx)
}

func (s *tariffService) ListActiveWithPrices(ctx context.Context) ([]domain.TariffWithPrices, error) {
	return s.repo.Tariffs.ListActiveWithPrices(ctx)
}
