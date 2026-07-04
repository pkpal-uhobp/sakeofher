package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/gateway"
	"sakeofher/internal/repository"
)

type subscriptionService struct {
	repo          *repository.Repositories
	remna         gateway.RemnawaveGateway
	notifications NotificationService
}

func NewSubscriptionService(repo *repository.Repositories, remna gateway.RemnawaveGateway, notifications NotificationService) SubscriptionService {
	return &subscriptionService{repo: repo, remna: remna, notifications: notifications}
}

func (s *subscriptionService) GetPublicByToken(ctx context.Context, token string) (*domain.PublicSubscription, error) {
	if strings.TrimSpace(token) == "" {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Subscriptions.GetPublicByToken(ctx, token)
}

func (s *subscriptionService) GetActiveByTelegramID(ctx context.Context, telegramID int64) (*domain.PublicSubscription, error) {
	if telegramID <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Subscriptions.GetActivePublicByTelegramID(ctx, telegramID)
}

func (s *subscriptionService) GetLatestByTelegramID(ctx context.Context, telegramID int64) (*domain.PublicSubscription, error) {
	if telegramID <= 0 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Subscriptions.GetLatestPublicByTelegramID(ctx, telegramID)
}

func (s *subscriptionService) ActivateAfterPayment(ctx context.Context, paymentID int64) error {
	payment, err := s.repo.Payments.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if payment.Status != domain.PaymentStatusPaid && payment.Status != domain.PaymentStatusActivationFailed {
		if payment.Status == domain.PaymentStatusActivated {
			return nil
		}

		return domain.ErrPaymentNotPaid
	}

	user, err := s.repo.Users.GetByID(ctx, payment.UserID)
	if err != nil {
		return err
	}

	squads, err := s.repo.Users.GetRemnaActiveSquads(ctx, user.ID)
	if err != nil {
		return err
	}
	if len(squads) == 0 {
		squads = defaultRemnaSquadsFromEnv()
	}

	now := time.Now()
	expiresAtPreview := now.AddDate(0, 0, payment.DurationDays)

	remnaUser, err := s.ensureRemnaUserWithSquads(ctx, user, payment.TrafficLimitBytes, expiresAtPreview, squads)
	if err != nil {
		_ = s.repo.Payments.MarkActivationFailed(ctx, paymentID, err)
		return err
	}

	return s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
		lockedUser, err := s.repo.Users.GetByIDForUpdate(ctx, user.ID)
		if err != nil {
			return err
		}

		if err := s.repo.Users.SetRemnaData(ctx, lockedUser.ID, domain.RemnaUserData{
			UUID:            remnaUser.UUID,
			Username:        remnaUser.Username,
			SubscriptionURL: remnaUser.SubscriptionURL,
			Status:          domain.RemnaStatus("active"),
		}); err != nil {
			return err
		}

		if len(squads) > 0 {
			if err := s.repo.Users.SetRemnaActiveSquads(ctx, lockedUser.ID, squads); err != nil {
				return err
			}
		}

		if err := s.createOrExtendSubscription(ctx, payment, now); err != nil {
			return err
		}

		if err := s.repo.Payments.MarkActivated(ctx, payment.ID, now); err != nil {
			return err
		}

		reqPayload, _ := json.Marshal(map[string]any{
			"payment_id": payment.ID,
			"user_id":    user.ID,
			"remna_uuid": remnaUser.UUID,
			"squads":     squads,
		})

		_ = s.repo.RemnaSync.Create(ctx, domain.RemnaSyncLog{
			UserID:         &user.ID,
			PaymentID:      &payment.ID,
			Action:         domain.RemnaSyncUpdateUser,
			Success:        true,
			RequestPayload: reqPayload,
		})

		return nil
	})
}

func (s *subscriptionService) ensureRemnaUser(ctx context.Context, user *domain.User, trafficLimitBytes int64, expiresAt time.Time) (*domain.RemnaUser, error) {
	squads, err := s.repo.Users.GetRemnaActiveSquads(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if len(squads) == 0 {
		squads = defaultRemnaSquadsFromEnv()
	}

	return s.ensureRemnaUserWithSquads(ctx, user, trafficLimitBytes, expiresAt, squads)
}

func (s *subscriptionService) createOrExtendSubscription(ctx context.Context, payment *domain.Payment, now time.Time) error {
	periodEnd := now.AddDate(0, 0, payment.PeriodDays)
	newExpire := now.AddDate(0, 0, payment.DurationDays)
	if periodEnd.After(newExpire) {
		periodEnd = newExpire
	}

	active, err := s.repo.Subscriptions.GetActiveByUserIDForUpdate(ctx, payment.UserID)
	if err == nil {
		base := active.ExpiresAt
		if base.Before(now) {
			base = now
		}

		active.LastPaymentID = &payment.ID
		active.TariffID = &payment.TariffID
		active.ExpiresAt = base.AddDate(0, 0, payment.DurationDays)
		active.TrafficLimitBytes = payment.TrafficLimitBytes
		active.PeriodStatus = domain.PeriodStatusActive
		active.Status = domain.SubscriptionStatusActive

		if active.CurrentPeriodEnd.Before(now) || active.CurrentPeriodEnd.After(active.ExpiresAt) {
			active.CurrentPeriodStart = now
			active.CurrentPeriodEnd = now.AddDate(0, 0, payment.PeriodDays)
			if active.CurrentPeriodEnd.After(active.ExpiresAt) {
				active.CurrentPeriodEnd = active.ExpiresAt
			}
		}

		return s.repo.Subscriptions.ExtendActive(ctx, active)
	}
	if err != domain.ErrNotFound {
		return err
	}

	sub := &domain.Subscription{
		UserID:             payment.UserID,
		TariffID:           &payment.TariffID,
		LastPaymentID:      &payment.ID,
		Status:             domain.SubscriptionStatusActive,
		StartedAt:          now,
		ExpiresAt:          newExpire,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		TrafficLimitBytes:  payment.TrafficLimitBytes,
		TrafficUsedBytes:   0,
		PeriodStatus:       domain.PeriodStatusActive,
	}

	return s.repo.Subscriptions.CreateActive(ctx, sub)
}

func (s *subscriptionService) DisableExpiredSubscriptions(ctx context.Context, limit int) error {
	now := time.Now()

	expired, err := s.repo.Subscriptions.FindExpiredActiveWithUsers(ctx, now, limit)
	if err != nil {
		return err
	}

	for _, item := range expired {
		publicItem := &domain.PublicSubscription{
			Subscription: item.Subscription,
			User:         item.User,
			Tariff:       item.Tariff,
		}

		if err := s.removeRemnaUserFromAllSquads(ctx, publicItem); err != nil {
			return err
		}

		deleteAfter := now.AddDate(0, 0, 7)

		if err := s.repo.Tx.WithinTransaction(ctx, func(ctx context.Context) error {
			if err := s.repo.Subscriptions.MarkExpired(ctx, item.Subscription.ID); err != nil {
				return err
			}

			if err := s.repo.Users.MarkRemnaDisabled(ctx, item.User.ID, now, deleteAfter); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		_ = s.notifications.Send(ctx, item.User.TelegramID, "Ваша подписка закончилась.\nДоступ временно отключён.\nПродлить можно в боте.")
	}

	return nil
}

func (s *subscriptionService) DeleteOldDisabledUsers(ctx context.Context, limit int) error {
	now := time.Now()

	users, err := s.repo.Users.FindDisabledReadyForDelete(ctx, now, limit)
	if err != nil {
		return err
	}

	for _, u := range users {
		if u.RemnaUUID != nil && *u.RemnaUUID != "" {
			if err := s.remna.DeleteUser(ctx, *u.RemnaUUID); err != nil {
				return err
			}
		}

		if err := s.repo.Users.MarkRemnaDeleted(ctx, u.ID, now); err != nil {
			return err
		}

		_ = s.notifications.Send(ctx, u.TelegramID, "Ваша старая подписка удалена из-за отсутствия продления более 7 дней.\nВы можете купить новую подписку.")
	}

	return nil
}

func (s *subscriptionService) SyncRemnaUsage(ctx context.Context, limit int) error {
	now := time.Now()

	items, err := s.repo.Subscriptions.FindActiveWithRemna(ctx, limit)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.User.RemnaUUID == nil || *item.User.RemnaUUID == "" {
			continue
		}

		traffic, err := s.remna.GetUserTraffic(ctx, *item.User.RemnaUUID)
		if err != nil {
			return err
		}

		used := traffic.UsedBytes
		limitBytes := item.Subscription.TrafficLimitBytes
		if traffic.LimitBytes > 0 {
			limitBytes = traffic.LimitBytes
		}

		if limitBytes > 0 && used >= limitBytes {
			if item.Subscription.PeriodStatus != domain.PeriodStatusTrafficExhausted {
				publicItem := &domain.PublicSubscription{
					Subscription: item.Subscription,
					User:         item.User,
					Tariff:       item.Tariff,
				}

				if err := s.removeRemnaUserFromAllSquads(ctx, publicItem); err != nil {
					return err
				}

				_ = s.notifications.Send(ctx, item.User.TelegramID, "Лимит трафика исчерпан.\nДоступ будет восстановлен после обновления периода или продления.")
			}

			if err := s.repo.Subscriptions.MarkTrafficExhausted(ctx, item.Subscription.ID, used, now); err != nil {
				return err
			}

			continue
		}

		if err := s.repo.Subscriptions.UpdateRemnaUsage(ctx, item.Subscription.ID, used, now); err != nil {
			return err
		}
	}

	return nil
}

func (s *subscriptionService) ResetTrafficPeriods(ctx context.Context, limit int) error {
	now := time.Now()

	items, err := s.repo.Subscriptions.FindReadyForTrafficReset(ctx, now, limit)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.User.RemnaUUID == nil || *item.User.RemnaUUID == "" {
			continue
		}

		if err := s.remna.ResetTraffic(ctx, *item.User.RemnaUUID); err != nil {
			return err
		}

		periodDays := item.Tariff.PeriodDays
		if periodDays <= 0 {
			periodDays = 30
		}

		nextEnd := now.AddDate(0, 0, periodDays)
		if nextEnd.After(item.Subscription.ExpiresAt) {
			nextEnd = item.Subscription.ExpiresAt
		}

		if item.Subscription.PeriodStatus == domain.PeriodStatusTrafficExhausted {
			publicItem := &domain.PublicSubscription{
				Subscription: item.Subscription,
				User:         item.User,
				Tariff:       item.Tariff,
			}

			_, err := s.restoreRemnaUserSquads(ctx, publicItem, nil)
			if err != nil {
				return err
			}
		}

		if err := s.repo.Subscriptions.ResetTrafficPeriod(ctx, item.Subscription.ID, now, nextEnd); err != nil {
			return err
		}
	}

	return nil
}

func (s *subscriptionService) PurchaseFromSite(ctx context.Context, input domain.SitePurchaseInput) (*domain.PublicSubscription, error) {
	return nil, domain.ErrInvalidInput
}

func (s *subscriptionService) RenewFromSite(ctx context.Context, input domain.SiteRenewInput) (*domain.PublicSubscription, error) {
	return nil, domain.ErrInvalidInput
}

func (s *subscriptionService) ensureRemnaUserForSite(ctx context.Context, user *domain.User, trafficLimitBytes int64, expiresAt time.Time) (*domain.RemnaUser, error) {
	return s.ensureRemnaUser(ctx, user, trafficLimitBytes, expiresAt)
}

func (s *subscriptionService) createOrExtendSiteSubscription(ctx context.Context, userID int64, tariff *domain.Tariff, trafficLimitBytes int64, now time.Time) error {
	tariffID := tariff.ID
	periodEnd := now.AddDate(0, 0, tariff.PeriodDays)
	newExpire := now.AddDate(0, 0, tariff.DurationDays)
	if periodEnd.After(newExpire) {
		periodEnd = newExpire
	}

	active, err := s.repo.Subscriptions.GetActiveByUserIDForUpdate(ctx, userID)
	if err == nil {
		base := active.ExpiresAt
		if base.Before(now) {
			base = now
		}

		active.TariffID = &tariffID
		active.LastPaymentID = nil
		active.ExpiresAt = base.AddDate(0, 0, tariff.DurationDays)
		active.TrafficLimitBytes = trafficLimitBytes
		active.PeriodStatus = domain.PeriodStatusActive
		active.Status = domain.SubscriptionStatusActive

		if active.CurrentPeriodEnd.Before(now) || active.CurrentPeriodEnd.After(active.ExpiresAt) {
			active.CurrentPeriodStart = now
			active.CurrentPeriodEnd = now.AddDate(0, 0, tariff.PeriodDays)
			if active.CurrentPeriodEnd.After(active.ExpiresAt) {
				active.CurrentPeriodEnd = active.ExpiresAt
			}
		}

		return s.repo.Subscriptions.ExtendActive(ctx, active)
	}
	if err != domain.ErrNotFound {
		return err
	}

	sub := &domain.Subscription{
		UserID:             userID,
		TariffID:           &tariffID,
		LastPaymentID:      nil,
		Status:             domain.SubscriptionStatusActive,
		StartedAt:          now,
		ExpiresAt:          newExpire,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		TrafficLimitBytes:  trafficLimitBytes,
		TrafficUsedBytes:   0,
		PeriodStatus:       domain.PeriodStatusActive,
	}

	return s.repo.Subscriptions.CreateActive(ctx, sub)
}

func remnaUsername(user *domain.User) string {
	candidates := make([]string, 0, 3)

	// Admin "Username" is Telegram alias and it must be the main Remnawave name.
	if user.TelegramUsername != nil && strings.TrimSpace(*user.TelegramUsername) != "" {
		candidates = append(candidates, *user.TelegramUsername)
	}

	if user.Alias != nil && strings.TrimSpace(*user.Alias) != "" {
		candidates = append(candidates, *user.Alias)
	}

	if user.RemnaUsername != nil && strings.TrimSpace(*user.RemnaUsername) != "" {
		oldValue := strings.TrimSpace(*user.RemnaUsername)
		if !strings.HasPrefix(oldValue, "tg_") {
			candidates = append(candidates, oldValue)
		}
	}

	for _, candidate := range candidates {
		normalized := normalizeSubscriptionRemnaUsername(candidate)
		if len(normalized) >= 3 {
			return normalized
		}
	}

	return fmt.Sprintf("tg_%d", user.TelegramID)
}

func normalizeSubscriptionRemnaUsername(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "@")
	if value == "" {
		return ""
	}

	var builder strings.Builder

	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
		case char >= 'A' && char <= 'Z':
			builder.WriteRune(char)
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
		case char == '_' || char == '-':
			builder.WriteRune(char)
		default:
			builder.WriteByte('_')
		}

		if builder.Len() >= 36 {
			break
		}
	}

	result := strings.Trim(builder.String(), "_-")
	if len(result) > 36 {
		result = result[:36]
	}

	return result
}
