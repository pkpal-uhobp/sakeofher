package app

import (
	"context"
	telegramtransport "sakeofher/internal/transport/telegram"
)

func RunBot(ctx context.Context) error {
	c, err := NewContainer(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	bot := telegramtransport.NewBot(c.Config.Telegram.BotToken, c.Services, c.Log)
	return bot.Run(ctx)
}
