package service

import (
	"context"
	"strings"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository"
)

type tariffService struct {
	repo *repository.Repositories
}

func NewTariffService(repo *repository.Repositories) TariffService {
	return &tariffService{repo: repo}
}

func (s *tariffService) ListActive(ctx context.Context) ([]domain.Tariff, error) {
	return s.repo.Tariffs.ListActive(ctx)
}

func (s *tariffService) ListActiveWithPrices(ctx context.Context) ([]domain.TariffWithPrices, error) {
	return s.repo.Tariffs.ListActiveWithPrices(ctx)
}

func (s *tariffService) ListAll(ctx context.Context) ([]domain.Tariff, error) {
	return s.repo.Tariffs.ListAll(ctx)
}

func (s *tariffService) GetByID(ctx context.Context, id int64) (*domain.Tariff, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Tariffs.GetAnyByID(ctx, id)
}

func (s *tariffService) Create(ctx context.Context, input domain.CreateTariffInput) (*domain.Tariff, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Title = strings.TrimSpace(input.Title)

	if input.Code == "" || input.Title == "" || input.DurationDays <= 0 || input.PeriodDays <= 0 || input.TrafficLimitGB <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Tariffs.Create(ctx, input)
}

func (s *tariffService) Update(ctx context.Context, id int64, input domain.UpdateTariffInput) (*domain.Tariff, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if input.Code != nil {
		value := strings.TrimSpace(*input.Code)
		input.Code = &value
	}
	if input.Title != nil {
		value := strings.TrimSpace(*input.Title)
		input.Title = &value
	}

	if input.DurationDays != nil && *input.DurationDays <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if input.PeriodDays != nil && *input.PeriodDays <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if input.TrafficLimitGB != nil && *input.TrafficLimitGB <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Tariffs.Update(ctx, id, input)
}

func (s *tariffService) Enable(ctx context.Context, id int64) (*domain.Tariff, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Tariffs.SetActive(ctx, id, true)
}

func (s *tariffService) Disable(ctx context.Context, id int64) (*domain.Tariff, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Tariffs.SetActive(ctx, id, false)
}
