package telegramtransport

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"

	"sakeofher/internal/domain"
)

func (b *Bot) siteSubscriptionURL(ctx context.Context, sub *domain.PublicSubscription) string {
	if sub == nil || sub.User.TelegramID <= 0 || b.services == nil || b.services.Site == nil {
		return ""
	}

	cfg, err := b.services.Site.GetConfig(ctx)
	if err != nil || cfg == nil {
		return ""
	}

	telegramID := strconv.FormatInt(sub.User.TelegramID, 10)
	pattern := strings.TrimSpace(cfg.SubscriptionURLPattern)
	if pattern != "" {
		return strings.ReplaceAll(pattern, "{telegram_id}", telegramID)
	}

	publicURL := strings.TrimRight(strings.TrimSpace(cfg.PublicURL), "/")
	secret := strings.Trim(strings.TrimSpace(cfg.SubscriptionPathSecret), "/")
	if publicURL == "" || secret == "" {
		return ""
	}

	return publicURL + "/" + secret + "/sub/" + telegramID
}

// Telegram rejects some technically valid local URLs in inline keyboard URL buttons
// (for example http://localhost:5173/...). Keep such links in message text only.
func telegramButtonURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" || host == "localhost" || host == "0.0.0.0" || host == "::1" {
		return ""
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsUnspecified() || ip.IsPrivate()) {
		return ""
	}
	return u.String()
}
