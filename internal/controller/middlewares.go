package controller

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller/logs"
)

func (bc *BotController) registerDefaultMiddlewares() {
	bot.WithMiddlewares(bc.writingActionMiddleware)(bc.bot)
}

func (bc *BotController) writingActionMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		_, err := bc.bot.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: update.Message.Chat.ID,
			Action: models.ChatActionTyping,
		})
		if err != nil {
			logs.LogBotErrorMessage(update, err)
			return
		}

		next(ctx, b, update)
	}
}
