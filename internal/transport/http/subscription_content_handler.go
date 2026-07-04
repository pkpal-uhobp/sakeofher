package httptransport

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/gateway/remnawave"
)

type remoteSubscriptionResponse struct {
	Body   string
	Header http.Header
}

func (h *PublicHandler) GetBase64SubscriptionByTelegramID(w http.ResponseWriter, r *http.Request) {
	subscriptionPath := strings.Trim(r.PathValue("subscription_path"), "/")
	if subscriptionPath != "" {
		if !h.isValidSubscriptionPath(subscriptionPath) {
			WriteDomainError(w, domain.ErrNotFound)
			return
		}
	}

	if subscriptionPath == "" {
		subscriptionPath = currentSubscriptionPathSecret()
	}

	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		WriteDomainError(w, domain.ErrInvalidInput)
		return
	}

	if wantsHTMLSubscriptionPage(r) {
		http.Redirect(
			w,
			r,
			fmt.Sprintf("/profile/%s/sub/%d", subscriptionPath, telegramID),
			http.StatusFound,
		)
		return
	}

	item, err := h.services.Subscriptions.GetLatestByTelegramID(r.Context(), telegramID)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	if isSubscriptionAccessExpired(item) {
		remote := h.makeExpiredSubscriptionResponse(r.Context(), r, item)
		copyRemoteResponseHeaders(w.Header(), remote.Header)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(remote.Body))
		return
	}

	remote, err := h.makeRemnawaveSubscriptionResponse(r.Context(), r, item)
	if err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}

	// Important for Happ:
	// sub-expire/sub-info are persistent app-management parameters. If they
	// were once sent for an expired subscription, Happ can keep displaying the
	// red renewal block until it receives explicit disabling parameters.
	remote.Body = stripExpiredHappMarkersFromActiveBody(remote.Body)

	copyRemoteResponseHeaders(w.Header(), remote.Header)
	ensureSubscriptionProfileHeaders(w.Header(), r, item)
	disableExpiredHappBlocksForActive(w.Header())

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	_, _ = w.Write([]byte(remote.Body))
}

func wantsHTMLSubscriptionPage(r *http.Request) bool {
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	switch format {
	case "base64", "sub", "subscription", "raw", "plain", "text":
		return false
	case "html", "page", "web":
		return true
	}

	accept := strings.ToLower(r.Header.Get("Accept"))
	userAgent := strings.ToLower(r.UserAgent())

	if strings.Contains(userAgent, "postman") ||
		strings.Contains(userAgent, "curl") ||
		strings.Contains(userAgent, "insomnia") ||
		strings.Contains(userAgent, "httpie") {
		return false
	}

	return strings.Contains(accept, "text/html")
}

func currentSubscriptionPathSecret() string {
	secret := strings.Trim(strings.TrimSpace(os.Getenv("SUBSCRIPTION_PATH_SECRET")), "/")
	if secret != "" {
		return secret
	}

	return "L0mENeiofHjdxC57"
}

func isSubscriptionAccessExpired(item *domain.PublicSubscription) bool {
	if item == nil {
		return true
	}

	if item.Subscription.ExpiresAt.Before(time.Now()) {
		return true
	}

	status := string(item.Subscription.Status)
	periodStatus := string(item.Subscription.PeriodStatus)

	return status == string(domain.SubscriptionStatusExpired) ||
		status == string(domain.SubscriptionStatusCancelled) ||
		periodStatus == string(domain.PeriodStatusFinished) ||
		periodStatus == string(domain.PeriodStatusTrafficExhausted)
}

func (h *PublicHandler) makeExpiredSubscriptionResponse(
	ctx context.Context,
	r *http.Request,
	item *domain.PublicSubscription,
) *remoteSubscriptionResponse {
	expiredHeaders := expiredSubscriptionHeaders(item)

	// Prefer Remnawave's own disabled/expired subscription body because it is
	// client-compatible. We still override/add our expired headers and bot URL.
	if remote, err := h.tryFetchRemnawaveExpiredSubscription(ctx, r, item); err == nil {
		for key, values := range expiredHeaders {
			remote.Header.Del(key)
			for _, value := range values {
				remote.Header.Add(key, value)
			}
		}

		return remote
	}

	// Fallback: empty Base64 body. Do not return comment-only Base64 body,
	// because Happ can try to parse such body as configs and show config error.
	return &remoteSubscriptionResponse{
		Body:   base64.StdEncoding.EncodeToString(nil),
		Header: expiredHeaders,
	}
}

func (h *PublicHandler) tryFetchRemnawaveExpiredSubscription(
	ctx context.Context,
	r *http.Request,
	item *domain.PublicSubscription,
) (*remoteSubscriptionResponse, error) {
	if item == nil {
		return nil, domain.ErrNotFound
	}

	remoteURL, err := h.resolveRemnawaveSubscriptionURL(ctx, item)
	if err != nil {
		return nil, err
	}

	remote, err := fetchRemoteSubscription(ctx, r, remoteURL)
	if err != nil {
		return nil, err
	}

	decoded := strings.ToLower(decodedSubscriptionText(remote.Body))
	if containsProxyNode(decoded) {
		return nil, fmt.Errorf("remnawave returned active nodes for expired subscription")
	}

	return remote, nil
}

func expiredSubscriptionHeaders(item *domain.PublicSubscription) http.Header {
	botURL := renewalBotURL()

	expireAt := time.Now().Add(-time.Hour).Unix()
	if item != nil && !item.Subscription.ExpiresAt.IsZero() {
		expireAt = item.Subscription.ExpiresAt.Unix()
	}

	headers := make(http.Header)
	headers.Set("profile-title", "SakeOfHer expired")
	headers.Set("profile-update-interval", "1")
	headers.Set("subscription-status", "expired")
	headers.Set("subscription-userinfo", fmt.Sprintf("upload=0; download=0; total=1; expire=%d", expireAt))
	headers.Set("profile-web-page-url", botURL)
	headers.Set("support-url", botURL)

	headers.Set("sub-expire", "1")
	headers.Set("sub-expire-button-link", botURL)
	headers.Set("sub-info-color", "red")
	headers.Set("sub-info-text", "Ваша подписка истекла. Продлите доступ в Telegram-боте.")
	headers.Set("sub-info-button-text", "Продлить")
	headers.Set("sub-info-button-link", botURL)

	if providerID := strings.TrimSpace(os.Getenv("HAPP_PROVIDER_ID")); providerID != "" {
		headers.Set("providerid", providerID)
	}

	return headers
}

func disableExpiredHappBlocksForActive(headers http.Header) {
	// Remove stale values copied from Remnawave or previous responses.
	for _, key := range []string{
		"sub-expire-button-link",
		"sub-info-color",
		"sub-info-button-text",
		"sub-info-button-link",
	} {
		headers.Del(key)
	}

	// Explicitly disable persistent Happ advanced blocks after renewal.
	// Happ docs: sub-expire is disabled by any value different from true/1,
	// and sub-info block is disabled by empty sub-info-text.
	headers.Set("sub-expire", "0")
	headers.Set("sub-info-text", "")
	headers.Set("subscription-status", "active")

	if providerID := strings.TrimSpace(os.Getenv("HAPP_PROVIDER_ID")); providerID != "" {
		headers.Set("providerid", providerID)
	}
}

func stripExpiredHappMarkersFromActiveBody(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return body
	}

	decoded, wasBase64 := tryDecodeSubscriptionBody(body)
	filtered := filterExpiredHappMetaLines(decoded)

	if wasBase64 {
		return base64.StdEncoding.EncodeToString([]byte(filtered))
	}

	return filtered
}

func tryDecodeSubscriptionBody(body string) (string, bool) {
	decoded, err := base64.StdEncoding.DecodeString(body)
	if err == nil {
		return string(decoded), true
	}

	return body, false
}

func filterExpiredHappMetaLines(decoded string) string {
	lines := strings.Split(decoded, "\n")
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		if isExpiredHappMetaLine(line) {
			continue
		}

		out = append(out, line)
	}

	return strings.TrimSpace(strings.Join(out, "\n"))
}

func isExpiredHappMetaLine(line string) bool {
	prepared := strings.TrimSpace(line)
	if prepared == "" {
		return false
	}

	for strings.HasPrefix(prepared, "#") {
		prepared = strings.TrimSpace(strings.TrimPrefix(prepared, "#"))
	}

	lower := strings.ToLower(prepared)

	for _, prefix := range []string{
		"sub-expire:",
		"sub-expire-button-link:",
		"sub-info-color:",
		"sub-info-text:",
		"sub-info-button-text:",
		"sub-info-button-link:",
		"subscription-status: expired",
	} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}

	return false
}

func decodedSubscriptionText(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	decoded, err := base64.StdEncoding.DecodeString(body)
	if err == nil {
		return string(decoded)
	}

	return body
}

func containsProxyNode(value string) bool {
	return strings.Contains(value, "vless://") ||
		strings.Contains(value, "vmess://") ||
		strings.Contains(value, "trojan://") ||
		strings.Contains(value, "hy2://") ||
		strings.Contains(value, "hysteria2://") ||
		strings.Contains(value, "ss://") ||
		strings.Contains(value, "socks://")
}

func renewalBotURL() string {
	for _, key := range []string{
		"TELEGRAM_BOT_URL",
		"BOT_URL",
		"PUBLIC_BOT_URL",
		"VITE_TELEGRAM_BOT_URL",
	} {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}

	username := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_USERNAME"))
	username = strings.TrimPrefix(username, "@")
	if username != "" {
		return "https://t.me/" + username
	}

	return "https://t.me/"
}

func (h *PublicHandler) makeRemnawaveSubscriptionResponse(
	ctx context.Context,
	sourceReq *http.Request,
	item *domain.PublicSubscription,
) (*remoteSubscriptionResponse, error) {
	if item == nil {
		return nil, domain.ErrNotFound
	}

	remoteURL, err := h.resolveRemnawaveSubscriptionURL(ctx, item)
	if err != nil {
		return nil, err
	}

	remote, err := fetchRemoteSubscription(ctx, sourceReq, remoteURL)
	if err != nil {
		return nil, fmt.Errorf("fetch remnawave subscription: %w", err)
	}

	remote.Body = strings.TrimSpace(remote.Body)
	if remote.Body == "" {
		return nil, fmt.Errorf("remnawave subscription is empty")
	}

	return remote, nil
}

func (h *PublicHandler) resolveRemnawaveSubscriptionURL(
	ctx context.Context,
	item *domain.PublicSubscription,
) (string, error) {
	if item.User.RemnaUUID != nil && strings.TrimSpace(*item.User.RemnaUUID) != "" {
		client := remnawave.NewClient(
			os.Getenv("REMNAWAVE_BASE_URL"),
			os.Getenv("REMNAWAVE_API_TOKEN"),
			15*time.Second,
		)

		remnaUser, err := client.GetUser(ctx, strings.TrimSpace(*item.User.RemnaUUID))
		if err == nil && strings.TrimSpace(remnaUser.SubscriptionURL) != "" {
			return strings.TrimSpace(remnaUser.SubscriptionURL), nil
		}
	}

	if item.SubscriptionURL != nil {
		if value := strings.TrimSpace(*item.SubscriptionURL); looksLikeRemoteSubscriptionURL(value) {
			return value, nil
		}
	}

	if item.User.SubscriptionURL != nil {
		if value := strings.TrimSpace(*item.User.SubscriptionURL); looksLikeRemoteSubscriptionURL(value) {
			return value, nil
		}
	}

	return "", fmt.Errorf("remnawave subscription URL is not available; create or update this user in Remnawave first")
}

func looksLikeRemoteSubscriptionURL(value string) bool {
	if value == "" {
		return false
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	host := strings.ToLower(parsed.Host)
	if host == "" ||
		strings.Contains(host, "localhost") ||
		strings.Contains(host, "127.0.0.1") {
		return false
	}

	publicURL := strings.TrimRight(strings.ToLower(os.Getenv("APP_PUBLIC_URL")), "/")
	if publicURL != "" && strings.HasPrefix(strings.ToLower(value), publicURL) {
		return false
	}

	return true
}

func fetchRemoteSubscription(ctx context.Context, sourceReq *http.Request, subscriptionURL string) (*remoteSubscriptionResponse, error) {
	remoteURL, err := url.Parse(subscriptionURL)
	if err != nil {
		return nil, err
	}

	query := remoteURL.Query()
	for key, values := range sourceReq.URL.Query() {
		if strings.EqualFold(key, "format") {
			continue
		}

		query.Del(key)
		for _, value := range values {
			query.Add(key, value)
		}
	}
	remoteURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remoteURL.String(), nil)
	if err != nil {
		return nil, err
	}

	copyClientRequestHeaders(req.Header, sourceReq.Header)

	userAgent := strings.TrimSpace(sourceReq.UserAgent())
	if userAgent == "" {
		userAgent = "sakeofher-subscription-fetcher/1.0"
	}

	req.Header.Set("Accept", "text/plain,*/*")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 25 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(raw))
	}

	return &remoteSubscriptionResponse{
		Body:   string(raw),
		Header: resp.Header.Clone(),
	}, nil
}

func copyClientRequestHeaders(dst http.Header, src http.Header) {
	for _, name := range []string{
		"Accept-Language",
		"Profile-Update-Interval",
		"X-HWID",
		"X-Hwid",
		"X-Device-OS",
		"X-Device-Model",
		"X-Ver-OS",
		"X-App-Version",
		"Hwid",
		"Device-OS",
	} {
		value := strings.TrimSpace(src.Get(name))
		if value != "" {
			dst.Set(name, value)
		}
	}
}

func copyRemoteResponseHeaders(dst http.Header, src http.Header) {
	for name, values := range src {
		if isHopByHopOrUnsafeResponseHeader(name) {
			continue
		}

		for _, value := range values {
			if strings.TrimSpace(value) == "" {
				continue
			}

			dst.Add(name, value)
		}
	}
}

func isHopByHopOrUnsafeResponseHeader(name string) bool {
	switch strings.ToLower(name) {
	case "connection",
		"keep-alive",
		"proxy-authenticate",
		"proxy-authorization",
		"te",
		"trailer",
		"transfer-encoding",
		"upgrade",
		"content-length",
		"content-encoding",
		"server",
		"date":
		return true
	default:
		return false
	}
}

func ensureSubscriptionProfileHeaders(headers http.Header, r *http.Request, item *domain.PublicSubscription) {
	if headers.Get("profile-title") == "" {
		headers.Set("profile-title", "SakeOfHer")
	}

	if headers.Get("profile-update-interval") == "" {
		interval := strings.TrimSpace(os.Getenv("SUBSCRIPTION_PROFILE_UPDATE_INTERVAL"))
		if interval == "" {
			interval = "4"
		}

		headers.Set("profile-update-interval", interval)
	}

	if headers.Get("profile-web-page-url") == "" {
		headers.Set("profile-web-page-url", publicProfileURL(r))
	}

	if headers.Get("subscription-userinfo") == "" && item != nil {
		headers.Set("subscription-userinfo", subscriptionUserInfoHeader(item))
	}
}

func publicProfileURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	if !strings.HasPrefix(path, "/profile/") {
		path = "/profile" + path
	}

	if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
		return scheme + "://" + forwardedHost + path
	}

	return scheme + "://" + r.Host + path
}

func subscriptionUserInfoHeader(item *domain.PublicSubscription) string {
	used := item.Subscription.TrafficUsedBytes
	limit := item.Subscription.TrafficLimitBytes
	expire := item.Subscription.ExpiresAt.Unix()

	if used < 0 {
		used = 0
	}

	if limit < 0 {
		limit = 0
	}

	return fmt.Sprintf("upload=0; download=%d; total=%d; expire=%d", used, limit, expire)
}
