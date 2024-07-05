package controller

import (
	"context"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller/logs"
)

func (bc *BotController) handlerWrapper(handler HandlerFunc, disableWebPagePreview bool) bot.HandlerFunc {
	return func(ctx context.Context, _ *bot.Bot, update *models.Update) {
		logs.LogBotIncomingMessage(update)

		user, err := bc.repo.GetOrCreateUser(ctx, update.Message.From.ID)
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

func (bc *BotController) registerHandler(
	pattern string,
	description string,
	disableWebPagePreview bool,
	handler HandlerFunc,
) {
	bc.bot.RegisterHandler(
		bot.HandlerTypeMessageText,
		pattern,
		bot.MatchTypePrefix,
		bc.handlerWrapper(handler, disableWebPagePreview),
	)

	var answer strings.Builder

	answer.WriteString(pattern)
	answer.WriteString(" - ")
	answer.WriteString(description)
	bc.commandList = append(bc.commandList, answer.String())
}
