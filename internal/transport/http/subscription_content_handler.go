package httptransport

import (
	"context"
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

	// One URL mode:
	//
	// Browser:
	//   Accept: text/html -> pretty page.
	//
	// App/Postman/curl:
	//   no text/html OR ?format=base64 -> raw Remnawave subscription body + headers.
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

	remote, err := h.makeRemnawaveSubscriptionResponse(r.Context(), r, item)
	if err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}

	copyRemoteResponseHeaders(w.Header(), remote.Header)
	ensureSubscriptionProfileHeaders(w.Header(), r, item)

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

	// Important:
	// Remnawave already generates the full subscription:
	// - all proxy configs,
	// - client routing metadata,
	// - profile title/update interval/userinfo headers,
	// - client-specific output by User-Agent.
	//
	// We return body and headers as close to Remnawave as possible.
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
	// These headers are what clients use for profile name, traffic display and auto-update.
	// If Remnawave returned them, keep Remnawave values untouched.
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

	if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
		return scheme + "://" + forwardedHost + "/profile" + strings.TrimSuffix(r.URL.Path, "/")
	}

	return scheme + "://" + r.Host + "/profile" + strings.TrimSuffix(r.URL.Path, "/")
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
