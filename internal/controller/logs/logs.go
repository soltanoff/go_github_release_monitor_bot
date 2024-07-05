package logs

import (
	"log/slog"

	"github.com/go-telegram/bot/models"
)

func LogBotIncomingMessage(update *models.Update) {
	slog.Info(
		"[BOT] New message",
		"chatID", update.Message.Chat.ID,
		"senderID", update.Message.From.ID,
		"username", update.Message.From.Username,
		"message", update.Message.Text,
	)
}

func LogBotOutgoingMessage(update *models.Update, answer string) {
	slog.Info(
		"[BOT] Send message",
		"chatID", update.Message.Chat.ID,
		"userID", update.Message.From.ID,
		"username", update.Message.From.Username,
		"message", answer,
	)
}

func LogBotErrorMessage(update *models.Update, err error) {
	slog.Error(
		"[BOT] User[%d|%d:@%s]: %s",
		"chatID", update.Message.Chat.ID,
		"userID", update.Message.From.ID,
		"username", update.Message.From.Username,
		"error", err.Error(),
	)
}
