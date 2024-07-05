package controller

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller/handlers"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/logs"
)

type BotController struct {
	bot                 *bot.Bot
	repo                *repo.Repository
	subscriptionHandler *handlers.SubscriptionsHandler
	commandList         []string
}

type HandlerFunc func(ctx context.Context, update *models.Update, user *entities.User) string

func New(
	telegramAPIKey string,
	repo *repo.Repository,
) (*BotController, error) {
	// bot.WithErrorsHandler(): логгировать ошибку на самом высоком уровне и/или использовать свой logger?
	b, err := bot.New(telegramAPIKey)
	if err != nil {
		logs.LogError("[BOT] Bot init error: %s", err.Error())
		return nil, fmt.Errorf("[BOT] failed to connect Telegram API: %w", err)
	}

	subscriptionHandler := handlers.New(repo)

	bc := BotController{bot: b, repo: repo, subscriptionHandler: subscriptionHandler}
	bc.registerDefaultMiddlewares()
	bc.registerDefaultHandler()
	bc.registerHandler(
		"/my_subscriptions",
		"view all subscriptions",
		true,
		subscriptionHandler.MySubscriptionsHandler,
	)
	bc.registerHandler(
		"/subscribe",
		"[github repo urls] subscribe to the new GitHub repository",
		true,
		subscriptionHandler.SubscribeHandler,
	)
	bc.registerHandler(
		"/unsubscribe",
		"[github repo urls] unsubscribe from the GitHub repository",
		true,
		subscriptionHandler.UnsubscribeHandler,
	)
	bc.registerHandler(
		"/remove_all_subscriptions",
		"remove all exists subscriptions",
		true,
		subscriptionHandler.RemoveAllSubscriptionsHandler,
	)

	return &bc, nil
}

func (bc *BotController) Start(ctx context.Context) {
	logs.LogInfo("[BOT] Starting bot...")
	bc.bot.Start(ctx)
	logs.LogInfo("[BOT] Close bot controller...")
}

func (bc *BotController) SendMessage(
	ctx context.Context,
	userExternalID int64,
	answer string,
	disableWebPagePreview bool,
) (err error) {
	_, err = bc.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:             userExternalID,
		Text:               answer,
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: &disableWebPagePreview},
	})
	if err != nil {
		return fmt.Errorf("[BOT] send message failed: %w", err)
	}

	logs.LogInfo("[BOT] <<< User %d: %s", userExternalID, answer)

	return nil
}
