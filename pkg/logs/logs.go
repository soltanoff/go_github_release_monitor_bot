package logs

import (
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot/models"
)

func LogInfo(message string, args ...any) {
	slog.Info(fmt.Sprintf(message, args...))
}

func LogWarn(message string, args ...any) {
	slog.Warn(fmt.Sprintf(message, args...))
}

func LogError(message string, args ...any) {
	slog.Error(fmt.Sprintf(message, args...))
}

func LogBotIncommingMessage(update *models.Update) {
	LogInfo(
		"User[%d|%d:@%s]: %s",
		update.Message.Chat.ID,
		update.Message.From.ID,
		update.Message.From.Username,
		update.Message.Text,
	)
}

func LogBotOutgoingMessage(update *models.Update, answer string) {
	LogInfo(
		"<<< User[%d|%d:@%s]: %s",
		update.Message.Chat.ID,
		update.Message.From.ID,
		update.Message.From.Username,
		answer,
	)
}

func LogBotErrorMessage(update *models.Update, err error) {
	LogError(
		"User[%d|%d:@%s]: %s",
		update.Message.Chat.ID,
		update.Message.From.ID,
		update.Message.From.Username,
		err.Error(),
	)
}
