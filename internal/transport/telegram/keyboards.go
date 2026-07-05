package telegramtransport

import (
	"fmt"

	"sakeofher/internal/domain"
)

func (b *Bot) mainMenuKeyboard(userID int64) *inlineKeyboardMarkup {
	rows := [][]inlineKeyboardButton{
		{{Text: "Купить VPN / Продлить", CallbackData: "main:connect"}},
		{{Text: "Личный кабинет", CallbackData: "main:profile"}},
		{{Text: "Как подключить", CallbackData: "main:instructions"}},
		{{Text: "Поддержка", CallbackData: "main:support"}},
	}
	if b.settings.isAdmin(userID) {
		rows = append(rows,
			[]inlineKeyboardButton{{Text: "Админ-панель", CallbackData: "admin:menu"}},
		)
	}
	return &inlineKeyboardMarkup{InlineKeyboard: rows}
}

func tariffsKeyboard(tariffs []domain.TariffWithPrices) *inlineKeyboardMarkup {
	rows := make([][]inlineKeyboardButton, 0, len(tariffs)+1)
	for _, item := range tariffs {
		rows = append(rows, []inlineKeyboardButton{{
			Text:         fmt.Sprintf("%s — %s / %s", durationLabel(item.DurationDays), trafficLabel(item.TrafficLimitBytes), tariffPriceShort(item)),
			CallbackData: fmt.Sprintf("tariff:%d", item.ID),
		}})
	}
	rows = append(rows, []inlineKeyboardButton{{Text: "В меню", CallbackData: "menu:home"}})
	return &inlineKeyboardMarkup{InlineKeyboard: rows}
}

func paymentMethodsKeyboard(item domain.TariffWithPrices, allowFree bool) *inlineKeyboardMarkup {
	rows := make([][]inlineKeyboardButton, 0, len(item.Prices)+3)
	for _, price := range item.Prices {
		if !price.IsActive {
			continue
		}
		switch price.Provider {
		case domain.PaymentProviderTelegramStars:
			if price.PaymentMethod == domain.PaymentMethodStars && price.StarsAmount != nil && *price.StarsAmount > 0 {
				rows = append(rows, []inlineKeyboardButton{{
					Text:         fmt.Sprintf("Telegram Stars — %d", *price.StarsAmount),
					CallbackData: fmt.Sprintf("pay:stars:%d", price.ID),
				}})
			}
		case domain.PaymentProviderCryptoBot:
			if price.PaymentMethod == domain.PaymentMethodCrypto {
				rows = append(rows, []inlineKeyboardButton{{
					Text:         fmt.Sprintf("CryptoBot — %s", priceAmountLabel(price)),
					CallbackData: fmt.Sprintf("pay:crypto:%d", price.ID),
				}})
			}
		}
	}
	if allowFree {
		rows = append(rows, []inlineKeyboardButton{{
			Text:         "Получить подписку без оплаты",
			CallbackData: fmt.Sprintf("free:activate:%d", item.ID),
		}})
	}
	rows = append(rows,
		[]inlineKeyboardButton{{Text: "Назад к тарифам", CallbackData: "main:connect"}},
		[]inlineKeyboardButton{{Text: "В меню", CallbackData: "menu:home"}},
	)
	return &inlineKeyboardMarkup{InlineKeyboard: rows}
}

func profileKeyboard(active bool, subURL string) *inlineKeyboardMarkup {
	rows := make([][]inlineKeyboardButton, 0, 6)
	if buttonURL := telegramButtonURL(subURL); buttonURL != "" {
		rows = append(rows, []inlineKeyboardButton{{Text: "Открыть ссылку подписки", URL: buttonURL}})
	}
	rows = append(rows,
		[]inlineKeyboardButton{{Text: "Обновить данные", CallbackData: "profile:refresh"}},
		[]inlineKeyboardButton{{Text: "Купить / Продлить", CallbackData: "profile:pay_options"}},
	)
	if active {
		rows = append(rows, []inlineKeyboardButton{{Text: "Как подключить", CallbackData: "main:instructions"}})
	}
	rows = append(rows, []inlineKeyboardButton{{Text: "В меню", CallbackData: "menu:home"}})
	return &inlineKeyboardMarkup{InlineKeyboard: rows}
}

func instructionsKeyboard(settings botSettings) *inlineKeyboardMarkup {
	return &inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
		{{Text: "iOS — Happ", URL: settings.IOSAppURL}},
		{{Text: "Android — Happ", URL: settings.AndroidAppURL}},
		{{Text: "Windows — Happ", URL: settings.WindowsAppURL}},
		{{Text: "В меню", CallbackData: "menu:home"}},
	}}
}

func supportKeyboard(settings botSettings) *inlineKeyboardMarkup {
	return &inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
		{{Text: "Написать в поддержку", URL: settings.SupportURL}},
		{{Text: "В меню", CallbackData: "menu:home"}},
	}}
}

func adminMenuKeyboard() *inlineKeyboardMarkup {
	return &inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
		{{Text: "Статистика", CallbackData: "admin:stats"}},
		{{Text: "Пользователи", CallbackData: "admin:users"}, {Text: "Подписки", CallbackData: "admin:subs"}},
		{{Text: "Тарифы", CallbackData: "admin:tariffs"}, {Text: "Stars", CallbackData: "admin:stars"}},
		{{Text: "Проверить TG ID", CallbackData: "admin:check"}},
		{{Text: "Выдать / продлить", CallbackData: "admin:grant"}},
		{{Text: "Продлить на дни", CallbackData: "admin:extend"}, {Text: "Изменить трафик", CallbackData: "admin:traffic"}},
		{{Text: "Включить", CallbackData: "admin:enable"}, {Text: "Отключить", CallbackData: "admin:disable"}},
		{{Text: "Отменить", CallbackData: "admin:cancel"}, {Text: "Alias", CallbackData: "admin:alias"}},
		{{Text: "Блок", CallbackData: "admin:block"}, {Text: "Разблок", CallbackData: "admin:unblock"}},
		{{Text: "Сообщение пользователю", CallbackData: "admin:message"}},
		{{Text: "Рассылка всем", CallbackData: "admin:broadcast"}},
		{{Text: "Сбросить действие", CallbackData: "admin:reset"}, {Text: "Закрыть", CallbackData: "admin:close"}},
	}}
}

func broadcastConfirmKeyboard() *inlineKeyboardMarkup {
	return &inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
		{{Text: "Отправить всем", CallbackData: "admin:broadcast_confirm"}},
		{{Text: "Отменить", CallbackData: "admin:broadcast_cancel"}},
		{{Text: "Админ-панель", CallbackData: "admin:menu"}},
	}}
}

func adminBackKeyboard() *inlineKeyboardMarkup {
	return &inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
		{{Text: "Админ-панель", CallbackData: "admin:menu"}},
		{{Text: "В меню", CallbackData: "menu:home"}},
	}}
}

func adminUserActionsKeyboard(tgID int64) *inlineKeyboardMarkup {
	return &inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
		{{Text: "Проверить снова", CallbackData: fmt.Sprintf("admin_user:check:%d", tgID)}},
		{{Text: "Включить", CallbackData: fmt.Sprintf("admin_user:enable:%d", tgID)}, {Text: "Отключить", CallbackData: fmt.Sprintf("admin_user:disable:%d", tgID)}},
		{{Text: "Выдать / продлить", CallbackData: fmt.Sprintf("admin_user:grant:%d", tgID)}, {Text: "Отменить", CallbackData: fmt.Sprintf("admin_user:cancel:%d", tgID)}},
		{{Text: "Админ-панель", CallbackData: "admin:menu"}},
	}}
}

func backToMenuKeyboard() *inlineKeyboardMarkup {
	return &inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
		{{Text: "В меню", CallbackData: "menu:home"}},
	}}
}
