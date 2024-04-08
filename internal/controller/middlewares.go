package controller

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/logs"
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

func (bc *BotController) handlerWrapper(handler HandlerFunc, disableWebPagePreview bool) bot.HandlerFunc {
	return func(ctx context.Context, _ *bot.Bot, update *models.Update) {
		logs.LogBotIncommingMessage(update)

		user, err := repo.GetOrCreateUser(ctx, update.Message.From.ID)
		if err != nil {
			logs.LogBotErrorMessage(update, err)
			return
		}

		answer := handler(ctx, update, &user)

		_, err = bc.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:             update.Message.Chat.ID,
			Text:               answer,
			ParseMode:          models.ParseModeHTML,
			ReplyParameters:    &models.ReplyParameters{MessageID: update.Message.ID},
			LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: &disableWebPagePreview},
		})
		if err != nil {
			logs.LogBotErrorMessage(update, err)
			return
		}

		logs.LogBotOutgoingMessage(update, answer)
	}
}
