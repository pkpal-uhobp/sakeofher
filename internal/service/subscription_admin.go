package service

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

func (s *subscriptionService) List(ctx context.Context, input domain.SubscriptionListInput) (*domain.SubscriptionListResponse, error) {
	if input.Status != "" &&
		input.Status != domain.SubscriptionStatusActive &&
		input.Status != domain.SubscriptionStatusExpired &&
		input.Status != domain.SubscriptionStatusCancelled {
		return nil, domain.ErrInvalidInput
	}

	items, total, err := s.repo.Subscriptions.ListPublic(ctx, input)
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

	return &domain.SubscriptionListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *subscriptionService) GetByID(ctx context.Context, id int64) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) CreateManual(ctx context.Context, input domain.CreateManualSubscriptionInput) (*domain.PublicSubscription, error) {
	if input.UserID <= 0 || input.TariffID <= 0 || input.TrafficLimitGB <= 0 {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.repo.Users.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	tariff, err := s.repo.Tariffs.GetByID(ctx, input.TariffID)
	if err != nil {
		return nil, err
	}

	trafficLimitBytes := domain.TrafficGBToBytes(input.TrafficLimitGB)
	now := time.Now()
	expiresAtPreview := now.AddDate(0, 0, tariff.DurationDays)

	remnaUser, err := s.ensureManualRemnaUser(ctx, user, trafficLimitBytes, expiresAtPreview, input.ActiveInternalSquads)
	if err != nil {
		return nil, err
	}

	var activeSub *domain.Subscription
	if err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.repo.Users.GetByIDForUpdate(ctx, input.UserID); err != nil {
			return err
		}

		if err := s.repo.Users.SetRemnaData(ctx, input.UserID, domain.RemnaUserData{
			UUID:            remnaUser.UUID,
			Username:        remnaUser.Username,
			SubscriptionURL: remnaUser.SubscriptionURL,
			Status:          domain.RemnaStatusActive,
		}); err != nil {
			return err
		}

		if err := s.createOrExtendSiteSubscription(ctx, input.UserID, tariff, trafficLimitBytes, now); err != nil {
			return err
		}

		activeSub, err = s.repo.Subscriptions.GetActiveByUserID(ctx, input.UserID)
		return err
	}); err != nil {
		return nil, err
	}

	if activeSub == nil {
		return nil, domain.ErrNotFound
	}

	siteURL := buildSiteSubscriptionURL(activeSub.PublicToken, user.TelegramID)
	if siteURL != "" {
		_ = s.repo.Users.SetRemnaData(ctx, input.UserID, domain.RemnaUserData{
			UUID:            remnaUser.UUID,
			Username:        remnaUser.Username,
			SubscriptionURL: siteURL,
			Status:          domain.RemnaStatusActive,
		})

		description := fmt.Sprintf("SakeOfHer user_id=%d subscription=%s", user.ID, siteURL)
		_, _ = s.remna.UpdateUser(ctx, domain.UpdateRemnaUserRequest{
			UUID:                 remnaUser.UUID,
			Description:          &description,
			ActiveInternalSquads: input.ActiveInternalSquads,
		})
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, activeSub.ID)
}

func (s *subscriptionService) Extend(ctx context.Context, id int64, input domain.ExtendSubscriptionInput) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	var out *domain.PublicSubscription

	if err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		sub, err := s.repo.Subscriptions.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}

		tariffID := int64(0)
		if sub.TariffID != nil {
			tariffID = *sub.TariffID
		}

		if input.TariffID != nil && *input.TariffID > 0 {
			tariffID = *input.TariffID
		}

		if tariffID <= 0 {
			return domain.ErrInvalidInput
		}

		tariff, err := s.repo.Tariffs.GetByID(ctx, tariffID)
		if err != nil {
			return err
		}

		days := tariff.DurationDays
		if input.Days != nil && *input.Days > 0 {
			days = *input.Days
		}

		now := time.Now()
		base := sub.ExpiresAt
		if base.Before(now) {
			base = now
		}

		nextExpiresAt := base.AddDate(0, 0, days)

		periodStart := sub.CurrentPeriodStart
		periodEnd := sub.CurrentPeriodEnd
		if periodEnd.Before(now) || periodEnd.After(nextExpiresAt) {
			periodStart = now
			periodEnd = now.AddDate(0, 0, tariff.PeriodDays)
			if periodEnd.After(nextExpiresAt) {
				periodEnd = nextExpiresAt
			}
		}

		if err := s.repo.Subscriptions.ExtendByID(ctx, id, tariff.ID, nextExpiresAt, periodStart, periodEnd, sub.TrafficLimitBytes); err != nil {
			return err
		}

		out, err = s.repo.Subscriptions.GetPublicByID(ctx, id)
		return err
	}); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *subscriptionService) Update(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if input.Status != nil {
		switch *input.Status {
		case domain.SubscriptionStatusActive, domain.SubscriptionStatusExpired, domain.SubscriptionStatusCancelled:
		default:
			return nil, domain.ErrInvalidInput
		}
	}

	if input.PeriodStatus != nil {
		switch *input.PeriodStatus {
		case domain.PeriodStatusActive, domain.PeriodStatusTrafficExhausted, domain.PeriodStatusFinished:
		default:
			return nil, domain.ErrInvalidInput
		}
	}

	if err := s.repo.Subscriptions.UpdateManual(ctx, id, input); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) UpdateTrafficLimit(ctx context.Context, id int64, input domain.UpdateTrafficLimitInput) (*domain.PublicSubscription, error) {
	if id <= 0 || input.TrafficLimitGB <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if err := s.repo.Subscriptions.UpdateTrafficLimit(ctx, id, domain.TrafficGBToBytes(input.TrafficLimitGB)); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) Disable(ctx context.Context, id int64) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if err := s.repo.Subscriptions.SetStatus(ctx, id, domain.SubscriptionStatusExpired, domain.PeriodStatusFinished); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) Enable(ctx context.Context, id int64) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if err := s.repo.Subscriptions.SetStatus(ctx, id, domain.SubscriptionStatusActive, domain.PeriodStatusActive); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) Cancel(ctx context.Context, id int64) (*domain.PublicSubscription, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	if err := s.repo.Subscriptions.SetStatus(ctx, id, domain.SubscriptionStatusCancelled, domain.PeriodStatusFinished); err != nil {
		return nil, err
	}

	return s.repo.Subscriptions.GetPublicByID(ctx, id)
}

func (s *subscriptionService) ensureManualRemnaUser(ctx context.Context, user *domain.User, trafficLimitBytes int64, expiresAt time.Time, activeInternalSquads []string) (*domain.RemnaUser, error) {
	username := manualRemnaUsername(user)
	description := fmt.Sprintf("SakeOfHer user_id=%d", user.ID)
	expiresAtUnix := expiresAt.Unix()
	telegramID := manualTelegramID(user.TelegramID)

	if user.RemnaUUID != nil && strings.TrimSpace(*user.RemnaUUID) != "" {
		remnaUser, err := s.remna.UpdateUser(ctx, domain.UpdateRemnaUserRequest{
			UUID:                 *user.RemnaUUID,
			Username:             username,
			Status:               "ACTIVE",
			TrafficLimitBytes:    &trafficLimitBytes,
			ExpiresAtUnix:        &expiresAtUnix,
			TrafficResetStrategy: "NO_RESET",
			Description:          &description,
			TelegramID:           telegramID,
			ActiveInternalSquads: activeInternalSquads,
		})
		if err != nil {
			return nil, err
		}

		if err := s.remna.EnableUser(ctx, *user.RemnaUUID); err != nil {
			return nil, err
		}

		return remnaUser, nil
	}

	return s.remna.CreateUser(ctx, domain.CreateRemnaUserRequest{
		Username:             username,
		TrafficLimitBytes:    trafficLimitBytes,
		ExpiresAtUnix:        expiresAtUnix,
		TrafficResetStrategy: "NO_RESET",
		Description:          description,
		TelegramID:           telegramID,
		ActiveInternalSquads: activeInternalSquads,
	})
}

func buildSiteSubscriptionURL(publicToken string, telegramID int64) string {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/")
	if base == "" {
		base = strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_PUBLIC_URL")), "/")
	}
	if base == "" {
		base = "http://localhost:5173"
	}

	secret := strings.Trim(strings.TrimSpace(os.Getenv("SUBSCRIPTION_PATH_SECRET")), "/")
	if secret == "" {
		secret = "L0mENeiofHjdxC57"
	}

	if telegramID > 0 {
		return fmt.Sprintf("%s/%s/sub/%d", base, secret, telegramID)
	}

	publicToken = strings.TrimSpace(publicToken)
	if publicToken == "" {
		return ""
	}

	return base + "/s/" + publicToken
}

func manualTelegramID(value int64) *int64 {
	if value <= 0 {
		return nil
	}

	return &value
}

var nonRemnaUsernameChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func manualRemnaUsername(user *domain.User) string {
	candidates := []string{
		stringPtrValue(user.TelegramUsername),
		stringPtrValue(user.Alias),
		stringPtrValue(user.TelegramFirstName),
		fmt.Sprintf("user_%d", user.ID),
	}

	for _, candidate := range candidates {
		value := strings.TrimSpace(candidate)
		if value == "" {
			continue
		}

		value = strings.TrimPrefix(value, "@")
		value = nonRemnaUsernameChars.ReplaceAllString(value, "_")
		value = strings.Trim(value, "_-")

		if len(value) > 36 {
			value = value[:36]
		}

		if len(value) >= 3 {
			return value
		}
	}

	return fmt.Sprintf("user_%d", user.ID)
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
