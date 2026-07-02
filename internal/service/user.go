package service

import (
	"context"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository"
)

type userService struct{ repo *repository.Repositories }

func NewUserService(repo *repository.Repositories) UserService { return &userService{repo: repo} }

func (s *userService) GetOrCreateTelegramUser(ctx context.Context, input domain.TelegramUserInput) (*domain.User, error) {
	if input.TelegramID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.Users.CreateOrUpdateTelegramUser(ctx, input)
}

func (s *userService) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	if telegramID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.Users.GetByTelegramID(ctx, telegramID)
}
