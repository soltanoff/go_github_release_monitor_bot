package controller

import (
	"context"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
)

func (bc *BotController) registerDefaultHandler() {
	bot.WithDefaultHandler(bc.handlerWrapper(bc.defaultHandler, true))(bc.bot)
	bc.registerHandler(
		"/start",
		"base command for user registration",
		false,
		bc.welcomeHandler,
	)
	bc.registerHandler(
		"/help",
		"view all commands",
		false,
		bc.welcomeHandler,
	)
}

func (bc *BotController) defaultHandler(_ context.Context, _ *models.Update, _ *entities.User) string {
	return "Say /help"
}

func (bc *BotController) welcomeHandler(_ context.Context, _ *models.Update, _ *entities.User) string {
	return strings.Join(bc.commandList, "\n")
}
