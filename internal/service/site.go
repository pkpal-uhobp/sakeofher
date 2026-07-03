package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/repository"
)

type siteService struct {
	repo                   *repository.Repositories
	botUsername            string
	publicURL              string
	subscriptionPathSecret string
}

func NewSiteService(repo *repository.Repositories, botUsername string, publicURL string, subscriptionPathSecret string) SiteService {
	return &siteService{
		repo:                   repo,
		botUsername:            normalizeBotUsername(botUsername),
		publicURL:              strings.TrimRight(strings.TrimSpace(publicURL), "/"),
		subscriptionPathSecret: strings.Trim(strings.TrimSpace(subscriptionPathSecret), "/"),
	}
}

func (s *siteService) GetConfig(ctx context.Context) (*domain.SiteConfig, error) {
	_ = ctx
	return &domain.SiteConfig{
		TelegramBotUsername:    s.botUsername,
		TelegramBotURL:         s.botURL(""),
		PaymentsLocation:       "telegram_bot",
		PublicURL:              s.publicURL,
		SubscriptionPathSecret: s.subscriptionPathSecret,
		SubscriptionURLPattern: s.subscriptionURLPattern(),
	}, nil
}

func (s *siteService) CreatePurchaseLink(ctx context.Context, input domain.SitePurchaseLinkInput) (*domain.SiteCheckoutLink, error) {
	if input.TariffID <= 0 || input.TrafficLimitGB <= 0 {
		return nil, domain.ErrInvalidInput
	}

	tariff, err := s.repo.Tariffs.GetByID(ctx, input.TariffID)
	if err != nil {
		return nil, err
	}
	if !tariff.IsActive {
		return nil, domain.ErrInvalidInput
	}

	payload := fmt.Sprintf("buy_t%d_g%d", tariff.ID, input.TrafficLimitGB)
	trafficBytes := domain.TrafficGBToBytes(input.TrafficLimitGB)
	now := time.Now()

	return &domain.SiteCheckoutLink{
		Action:               domain.SiteCheckoutActionPurchase,
		StartPayload:         payload,
		TelegramBotURL:       s.botURL(payload),
		TelegramBotUsername:  s.botUsername,
		Tariff:               *tariff,
		TrafficLimitGB:       input.TrafficLimitGB,
		TrafficLimitBytes:    trafficBytes,
		NextExpiresAtPreview: now.AddDate(0, 0, tariff.DurationDays),
		Note:                 "Оплата и окончательная активация подписки выполняются в Telegram-боте. Сайт только формирует параметры покупки.",
	}, nil
}

func (s *siteService) CreateRenewLink(ctx context.Context, input domain.SiteRenewLinkInput) (*domain.SiteCheckoutLink, error) {
	token := strings.TrimSpace(input.PublicToken)
	if token == "" {
		return nil, domain.ErrInvalidInput
	}

	current, err := s.repo.Subscriptions.GetPublicByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	tariff := current.Tariff
	if current.Subscription.TariffID != nil && *current.Subscription.TariffID > 0 {
		loaded, err := s.repo.Tariffs.GetByID(ctx, *current.Subscription.TariffID)
		if err == nil && loaded.IsActive {
			tariff = *loaded
		}
	}

	trafficBytes := current.Subscription.TrafficLimitBytes
	if trafficBytes <= 0 {
		trafficBytes = tariff.TrafficLimitBytes
	}
	trafficGB := domain.TrafficBytesToGB(trafficBytes)
	if trafficGB <= 0 {
		trafficGB = domain.TrafficBytesToGB(tariff.TrafficLimitBytes)
		trafficBytes = tariff.TrafficLimitBytes
	}

	payload := fmt.Sprintf("renew_%s", token)
	base := time.Now()
	if current.Subscription.ExpiresAt.After(base) {
		base = current.Subscription.ExpiresAt
	}

	return &domain.SiteCheckoutLink{
		Action:               domain.SiteCheckoutActionRenew,
		StartPayload:         payload,
		TelegramBotURL:       s.botURL(payload),
		TelegramBotUsername:  s.botUsername,
		Tariff:               tariff,
		TrafficLimitGB:       trafficGB,
		TrafficLimitBytes:    trafficBytes,
		PublicToken:          token,
		CurrentExpiresAt:     &current.Subscription.ExpiresAt,
		NextExpiresAtPreview: base.AddDate(0, 0, tariff.DurationDays),
		Note:                 "Продление оплачивается в Telegram-боте. Лимит трафика подтянут автоматически из текущей подписки.",
	}, nil
}

func (s *siteService) subscriptionURLPattern() string {
	if s.publicURL == "" || s.subscriptionPathSecret == "" {
		return ""
	}
	return s.publicURL + "/" + s.subscriptionPathSecret + "/sub/{telegram_id}"
}

func (s *siteService) botURL(payload string) string {
	username := s.botUsername
	if username == "" {
		return ""
	}
	base := "https://t.me/" + username
	if payload == "" {
		return base
	}
	return base + "?start=" + url.QueryEscape(payload)
}

func normalizeBotUsername(username string) string {
	username = strings.TrimSpace(username)
	username = strings.TrimPrefix(username, "@")
	return username
}
