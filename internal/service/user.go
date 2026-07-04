package service

import (
	"context"
	"os"
	"strings"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/gateway/remnawave"
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
		Items:  items,
		Total:  total,
		Limit:  limit,
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
		value = strings.TrimPrefix(value, "@")
		input.TelegramUsername = &value
	}

	if input.Alias != nil {
		value := strings.TrimSpace(*input.Alias)
		input.Alias = &value
	}

	updated, err := s.repo.Users.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	if updated.RemnaUUID != nil && strings.TrimSpace(*updated.RemnaUUID) != "" {
		username := strings.TrimSpace(stringValue(updated.TelegramUsername))
		if username == "" {
			username = strings.TrimSpace(stringValue(updated.Alias))
		}

		if username != "" {
			_, _ = remnaClientFromEnv().UpdateUser(ctx, domain.UpdateRemnaUserRequest{
				UUID:     *updated.RemnaUUID,
				Username: normalizeUserRemnaUsername(username, updated.ID),
			})
		}
	}

	return updated, nil
}

func (s *userService) Block(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.repo.Users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.RemnaUUID != nil && strings.TrimSpace(*user.RemnaUUID) != "" {
		if err := remnaClientFromEnv().DisableUser(ctx, *user.RemnaUUID); err != nil {
			return nil, err
		}
	}

	return s.repo.Users.SetStatus(ctx, id, domain.UserStatusBlocked)
}

func (s *userService) Unblock(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.repo.Users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.RemnaUUID != nil && strings.TrimSpace(*user.RemnaUUID) != "" {
		if err := remnaClientFromEnv().EnableUser(ctx, *user.RemnaUUID); err != nil {
			return nil, err
		}
	}

	return s.repo.Users.SetStatus(ctx, id, domain.UserStatusActive)
}

func (s *userService) MarkDeleted(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.repo.Users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.RemnaUUID != nil && strings.TrimSpace(*user.RemnaUUID) != "" {
		if err := remnaClientFromEnv().DeleteUser(ctx, *user.RemnaUUID); err != nil {
			return nil, err
		}
	}

	return s.repo.Users.MarkDeleted(ctx, id, time.Now())
}

func remnaClientFromEnv() *remnawave.Client {
	return remnawave.NewClient(
		os.Getenv("REMNAWAVE_BASE_URL"),
		os.Getenv("REMNAWAVE_API_TOKEN"),
		15*time.Second,
	)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func normalizeUserRemnaUsername(value string, id int64) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "@")
	value = nonRemnaUsernameChars.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_-")

	if len(value) > 36 {
		value = value[:36]
	}

	if len(value) < 3 {
		return "user_" + strings.TrimPrefix(strings.TrimSpace(time.Now().Format("150405")), "-")
	}

	return value
}
