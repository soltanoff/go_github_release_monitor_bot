package utils

import (
	"fmt"
	"github.com/go-telegram/bot/models"
	"log/slog"
)

func LogIncommingMessage(update *models.Update) {
	slog.Info(fmt.Sprintf(
		"User[%d|%d:@%s]: %s",
		update.Message.Chat.ID,
		update.Message.From.ID,
		update.Message.From.Username,
		update.Message.Text,
	))
}

func LogOutgoingMessage(update *models.Update, answer string) {
	slog.Info(fmt.Sprintf(
		"<<< User[%d|%d:@%s]: %s",
		update.Message.Chat.ID,
		update.Message.From.ID,
		update.Message.From.Username,
		answer,
	))
}
