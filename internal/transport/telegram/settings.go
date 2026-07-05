package telegramtransport

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type botSettings struct {
	SupportURL         string
	IOSAppURL          string
	AndroidAppURL      string
	WindowsAppURL      string
	AdminIDs           map[int64]struct{}
	AllowFreePurchase  bool
	DefaultRemnaSquads []string
	BroadcastDelay     time.Duration
}

func loadBotSettings() botSettings {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	_ = appEnv

	// Production-safe default: users never receive a subscription without payment.
	// Enable only manually for local one-off debugging, not for a real bot.
	allowFree := envBool("BOT_ALLOW_FREE_PURCHASE", envBool("BOT_ALLOW_DEV_ACTIVATION", false))

	return botSettings{
		SupportURL:         envString("SUPPORT_URL", "https://t.me/username"),
		IOSAppURL:          envString("IOS_APP_URL", "https://apps.apple.com/us/app/happ-proxy-utility/id6504287215"),
		AndroidAppURL:      envString("ANDROID_APP_URL", "https://play.google.com/store/apps/details?id=com.happproxy"),
		WindowsAppURL:      envString("WINDOWS_APP_URL", "https://github.com/Happ-proxy/happ-desktop/releases/latest/download/setup-Happ.x86.exe"),
		AdminIDs:           parseAdminIDs(envString("TELEGRAM_ADMIN_IDS", os.Getenv("ADMIN_IDS"))),
		AllowFreePurchase:  allowFree,
		DefaultRemnaSquads: parseCSV(envString("BOT_REMNAWAVE_INTERNAL_SQUADS", os.Getenv("REMNAWAVE_DEFAULT_INTERNAL_SQUADS"))),
		BroadcastDelay:     time.Duration(envInt("BOT_BROADCAST_DELAY_MS", 80)) * time.Millisecond,
	}
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func parseAdminIDs(raw string) map[int64]struct{} {
	result := make(map[int64]struct{})
	for _, part := range parseCSV(raw) {
		id, err := strconv.ParseInt(part, 10, 64)
		if err == nil && id > 0 {
			result[id] = struct{}{}
		}
	}
	return result
}

func parseCSV(raw string) []string {
	items := strings.Split(raw, ",")
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, part := range items {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		result = append(result, part)
	}
	return result
}

func (s botSettings) isAdmin(telegramID int64) bool {
	_, ok := s.AdminIDs[telegramID]
	return ok
}
