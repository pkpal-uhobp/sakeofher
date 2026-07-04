package service

import (
	"context"
	"strings"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository"
)

type userService struct {
	repo *repository.Repositories
}

func NewUserService(repo *repository.Repositories) UserService {
	return &userService{repo: repo}
}

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

func (s *userService) List(ctx context.Context, input domain.UserListInput) (*domain.UserListResponse, error) {
	if input.Status != "" && input.Status != domain.UserStatusActive && input.Status != domain.UserStatusBlocked && input.Status != domain.UserStatusDeleted {
		return nil, domain.ErrInvalidInput
	}

	items, total, err := s.repo.Users.List(ctx, input)
	if err != nil {
		return nil, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	return &domain.UserListResponse{
		Items: items,
		Total: total,
		Limit: limit,
		Offset: offset,
	}, nil
}

func (s *userService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Users.GetByID(ctx, id)
}

func (s *userService) Update(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if input.Status != nil {
		switch *input.Status {
		case domain.UserStatusActive, domain.UserStatusBlocked, domain.UserStatusDeleted:
		default:
			return nil, domain.ErrInvalidInput
		}
	}

	if input.TelegramUsername != nil {
		value := strings.TrimSpace(*input.TelegramUsername)
		input.TelegramUsername = &value
	}
	if input.Alias != nil {
		value := strings.TrimSpace(*input.Alias)
		input.Alias = &value
	}

	return s.repo.Users.Update(ctx, id, input)
}

func (s *userService) Block(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Users.SetStatus(ctx, id, domain.UserStatusBlocked)
}

func (s *userService) Unblock(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Users.SetStatus(ctx, id, domain.UserStatusActive)
}

func (s *userService) MarkDeleted(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Users.MarkDeleted(ctx, id, time.Now())
}
