package service

import (
	"context"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository"
)

type UserService struct{ repo *repository.Repositories }

func NewUserService(repo *repository.Repositories) *UserService { return &UserService{repo: repo} }

func (s *UserService) GetOrCreateTelegramUser(ctx context.Context, input domain.TelegramUserInput) (*domain.User, error) {
	return s.repo.Users.CreateTelegramUser(ctx, input)
}
