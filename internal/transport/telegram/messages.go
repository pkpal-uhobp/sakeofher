package telegramtransport

import (
	"fmt"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

func welcomeMessage(firstName string) string {
	name := strings.TrimSpace(firstName)
	if name != "" {
		name = ", " + name
	}
	return strings.TrimSpace(fmt.Sprintf(`Привет%s

Добро пожаловать в SakeOfHer VPN.

Здесь можно купить или продлить подписку, открыть личный кабинет, проверить срок и трафик, получить ссылку подписки и инструкцию по подключению.

Выберите нужный пункт ниже:`, name))
}

func helpMessage(isAdmin bool) string {
	text := strings.TrimSpace(`Доступные действия:

/start — главное меню
/status — личный кабинет и ссылка подписки
/renew — продлить подписку
/help — помощь

В меню доступны разделы:
• Купить VPN / Продлить
• Личный кабинет
• Как подключить
• Поддержка`)
	if isAdmin {
		text += strings.TrimSpace(`

Админ-команды:
/admin — админ-панель
/check TG_ID — проверить пользователя
/grant TG_ID [tariff_id] [alias] — выдать или продлить подписку
/extend TG_ID DAYS — продлить на дни
/traffic TG_ID GB — изменить лимит трафика
/enable TG_ID — включить подписку
/disable TG_ID — отключить подписку
/cancel TG_ID — отменить подписку
/block TG_ID — заблокировать пользователя
/unblock TG_ID — разблокировать пользователя
/alias TG_ID alias — изменить alias
/msg TG_ID text — написать пользователю
/users [query] — пользователи
/subs [all|active|expired|cancelled] — подписки
/admin_tariffs — тарифы
/stars — баланс Telegram Stars
/broadcast text — рассылка всем активным пользователям`)
	}
	return text
}

func adminMenuMessage() string {
	return strings.TrimSpace(`Админ-панель

Что уже можно делать из бота:
• смотреть статистику, пользователей, подписки и тарифы;
• проверять пользователя по Telegram ID;
• вручную выдавать и продлевать подписку;
• включать, отключать и отменять подписку;
• менять лимит трафика;
• блокировать и разблокировать пользователя;
• менять alias и писать пользователю;
• делать рассылку всем активным пользователям с предпросмотром и подтверждением.

Выберите действие кнопками ниже или используйте команды из /help.`)
}

func tariffsMessage(tariffs []domain.TariffWithPrices) string {
	if len(tariffs) == 0 {
		return "Активных тарифов пока нет. Добавь тарифы в админке или через seed-миграции."
	}
	var b strings.Builder
	b.WriteString("Выберите тариф:\n\n")
	for _, item := range tariffs {
		b.WriteString(fmt.Sprintf("• %s — %s, %s\n", item.Title, durationLabel(item.DurationDays), trafficLabel(item.TrafficLimitBytes)))
		b.WriteString(fmt.Sprintf("Оплата: %s\n", tariffPricesLabel(item)))
		if item.Description != nil && strings.TrimSpace(*item.Description) != "" {
			b.WriteString("")
			b.WriteString(strings.TrimSpace(*item.Description))
			b.WriteByte('\n')
		}
	}
	b.WriteString("\nПосле выбора тарифа выберите доступный способ оплаты.")
	return b.String()
}

func tariffPaymentMessage(item domain.TariffWithPrices) string {
	return fmt.Sprintf(strings.TrimSpace(`Тариф: %s
Срок: %s
Трафик: %s

Варианты оформления:
%s

Выберите действие ниже.`),
		item.Title,
		durationLabel(item.DurationDays),
		trafficLabel(item.TrafficLimitBytes),
		tariffPricesMultiline(item),
	)
}

func cryptoNotReadyMessage(provider string) string {
	return fmt.Sprintf(strings.TrimSpace(`%s пока не подключён в Go-боте полностью.

Меню и тарифы уже готовы. Для Telegram Stars уже есть invoice-flow. Для CryptoBot/Tribute нужен следующий шаг: создание invoice через gateway + webhook/polling подтверждения.

Пока можно написать в поддержку.`), provider)
}

func invoiceCreatedMessage() string {
	return "Счёт Telegram Stars создан. Нажмите кнопку оплаты в сообщении ниже."
}

func paymentActivatedMessage(sub *domain.PublicSubscription, siteURL string) string {
	var b strings.Builder
	b.WriteString("Оплата подтверждена. Подписка активирована.\n\n")
	b.WriteString(publicSubscriptionText(sub))
	if siteURL != "" {
		b.WriteString("\n\nСсылка подписки:\n")
		b.WriteString(siteURL)
	} else {
		b.WriteString("\n\nСсылка сайта не настроена: проверь APP_PUBLIC_URL и SUBSCRIPTION_PATH_SECRET.")
	}
	return b.String()
}

func freeActivatedMessage(sub *domain.PublicSubscription, siteURL string) string {
	var b strings.Builder
	b.WriteString("Подписка активирована без оплаты.\n\n")
	b.WriteString(publicSubscriptionText(sub))
	if siteURL != "" {
		b.WriteString("\n\nСсылка подписки:\n")
		b.WriteString(siteURL)
	} else {
		b.WriteString("\n\nСсылка сайта не настроена: проверь APP_PUBLIC_URL и SUBSCRIPTION_PATH_SECRET.")
	}
	return b.String()
}

func statusMessage(sub *domain.PublicSubscription, siteURL string) string {
	if sub == nil || sub.Subscription.ID == 0 {
		return "У вас пока нет подписки. Выберите тариф в разделе «Купить VPN / Продлить»."
	}
	var b strings.Builder
	b.WriteString(publicSubscriptionText(sub))
	if siteURL != "" {
		b.WriteString("\n\nСсылка подписки:\n")
		b.WriteString(siteURL)
	} else {
		b.WriteString("\n\nСсылка сайта не настроена: проверь APP_PUBLIC_URL и SUBSCRIPTION_PATH_SECRET.")
	}
	return b.String()
}

func publicSubscriptionText(sub *domain.PublicSubscription) string {
	if sub == nil {
		return "Подписка не найдена."
	}
	status := "активна"
	if sub.Subscription.Status != domain.SubscriptionStatusActive || time.Now().After(sub.Subscription.ExpiresAt) {
		status = "неактивна"
	}
	periodStatus := string(sub.Subscription.PeriodStatus)
	left := time.Until(sub.Subscription.ExpiresAt)
	leftText := "истекла"
	if left > 0 {
		leftText = humanDuration(left)
	}
	usedGB := domain.TrafficBytesToGB(sub.Subscription.TrafficUsedBytes)
	limitGB := domain.TrafficBytesToGB(sub.Subscription.TrafficLimitBytes)
	if limitGB <= 0 && sub.Tariff.TrafficLimitBytes > 0 {
		limitGB = domain.TrafficBytesToGB(sub.Tariff.TrafficLimitBytes)
	}
	return fmt.Sprintf(strings.TrimSpace(`Личный кабинет

Подписка: %s
Тариф: %s
Период: %s
Активна до: %s
Осталось: %s
Трафик: %d / %d ГБ`),
		status,
		valueOrDash(sub.Tariff.Title),
		periodStatus,
		sub.Subscription.ExpiresAt.Format("02.01.2006 15:04"),
		leftText,
		usedGB,
		limitGB,
	)
}

func instructionsMessage() string {
	return strings.TrimSpace(`Как подключить VPN:

1) Оформите или продлите подписку.
2) Откройте «Личный кабинет» и скопируйте ссылку подписки.
3) Установите приложение Happ для вашей платформы.
4) В Happ выберите Import subscription / Импорт по ссылке.
5) Вставьте ссылку, сохраните профиль и включите VPN.

Если не подключается: обновите подписку в приложении и перезапустите VPN.`)
}

func supportMessage(settings botSettings) string {
	return "Поддержка: " + settings.SupportURL
}

func durationLabel(days int) string {
	switch days {
	case 30:
		return "1 месяц"
	case 60:
		return "2 месяца"
	case 90:
		return "3 месяца"
	}
	if days%30 == 0 && days > 0 {
		return fmt.Sprintf("%d мес.", days/30)
	}
	return fmt.Sprintf("%d дн.", days)
}

func moneyRubMinor(amountMinor *int64) string {
	if amountMinor == nil || *amountMinor <= 0 {
		return "цена не указана"
	}
	return fmt.Sprintf("%d ₽", *amountMinor/100)
}

func trafficLabel(bytes int64) string {
	gb := domain.TrafficBytesToGB(bytes)
	if gb <= 0 {
		return "без лимита"
	}
	return fmt.Sprintf("%d ГБ", gb)
}

func humanDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%d дн.", days)
	}
	hours := int(d.Hours())
	if hours > 0 {
		return fmt.Sprintf("%d ч.", hours)
	}
	minutes := int(d.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%d мин.", minutes)
	}
	return "меньше минуты"
}

func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return strings.TrimSpace(value)
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func tariffPriceShort(item domain.TariffWithPrices) string {
	for _, price := range item.Prices {
		if !price.IsActive || price.Provider == domain.PaymentProviderTribute {
			continue
		}
		if price.Provider == domain.PaymentProviderTelegramStars && price.StarsAmount != nil && *price.StarsAmount > 0 {
			return fmt.Sprintf("%d Stars", *price.StarsAmount)
		}
	}

	for _, price := range item.Prices {
		if !price.IsActive || price.Provider == domain.PaymentProviderTribute {
			continue
		}
		if price.AmountMinor != nil && *price.AmountMinor > 0 {
			return moneyRubMinor(price.AmountMinor)
		}
	}

	if item.PriceRub > 0 {
		return fmt.Sprintf("%d ₽", item.PriceRub)
	}

	return "цена не указана"
}

func tariffPricesLabel(item domain.TariffWithPrices) string {
	labels := make([]string, 0, len(item.Prices))
	for _, price := range item.Prices {
		if !price.IsActive || price.Provider == domain.PaymentProviderTribute {
			continue
		}

		label := priceLabel(price)
		if label == "" {
			continue
		}

		labels = append(labels, label)
	}

	if len(labels) == 0 {
		return "нет активных способов оплаты"
	}

	return strings.Join(labels, ", ")
}

func tariffPricesMultiline(item domain.TariffWithPrices) string {
	lines := make([]string, 0, len(item.Prices))
	for _, price := range item.Prices {
		if !price.IsActive || price.Provider == domain.PaymentProviderTribute {
			continue
		}

		label := priceLabel(price)
		if label == "" {
			continue
		}

		lines = append(lines, "• "+label)
	}

	if len(lines) == 0 {
		return "• нет активных способов оплаты"
	}

	return strings.Join(lines, "\n")
}

func priceLabel(price domain.TariffPrice) string {
	switch price.Provider {
	case domain.PaymentProviderTelegramStars:
		if price.StarsAmount != nil {
			return fmt.Sprintf("Telegram Stars — %d", *price.StarsAmount)
		}
		return "Telegram Stars"
	case domain.PaymentProviderCryptoBot:
		return "CryptoBot — " + priceAmountLabel(price)
	case domain.PaymentProviderTribute:
		return ""
	default:
		return string(price.Provider) + " — " + priceAmountLabel(price)
	}
}

func priceAmountLabel(price domain.TariffPrice) string {
	if price.StarsAmount != nil && *price.StarsAmount > 0 {
		return fmt.Sprintf("%d Stars", *price.StarsAmount)
	}
	if price.AmountMinor != nil && *price.AmountMinor > 0 {
		if strings.EqualFold(price.Currency, "RUB") {
			return moneyRubMinor(price.AmountMinor)
		}
		return fmt.Sprintf("%d %s", *price.AmountMinor, price.Currency)
	}
	return "цена не указана"
}
