package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/config"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller/handlers"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
)

type BotController struct {
	bot                 *bot.Bot
	repository          *repo.Repository
	subscriptionHandler *handlers.SubscriptionsHandler
	commandList         []string
}

type HandlerFunc func(ctx context.Context, update *models.Update, user *entities.User) string

func NewBotController(
	cfg *config.Config,
	repository *repo.Repository,
) (*BotController, error) {
	// bot.WithErrorsHandler(): логгировать ошибку на самом высоком уровне и/или использовать свой logger?
	b, err := bot.New(cfg.TelegramAPIKey)
	if err != nil {
		slog.Error("[BOT] Bot init error", "error", err.Error())

		return nil, fmt.Errorf("[BOT] failed to connect Telegram API: %w", err)
	}

	subscriptionHandler := handlers.NewSubscriptionsHandler(repository)

	bc := BotController{bot: b, repository: repository, subscriptionHandler: subscriptionHandler}
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
		"[github repository urls] subscribe to the new GitHub repository",
		true,
		subscriptionHandler.SubscribeHandler,
	)
	bc.registerHandler(
		"/unsubscribe",
		"[github repository urls] unsubscribe from the GitHub repository",
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
	slog.Info("[BOT] Starting bot...")
	bc.bot.Start(ctx)
	slog.Info("[BOT] Close bot controller...")
}

func (bc *BotController) SendMessage(
	ctx context.Context,
	userExternalID int64,
	answer string,
	disableWebPagePreview bool,
) error {
	_, err := bc.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:             userExternalID,
		Text:               answer,
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: &disableWebPagePreview},
	})
	if err != nil {
		return fmt.Errorf("[BOT] send message failed: %w", err)
	}

	slog.Info("[BOT] <<< ", "receiverID", userExternalID, "answer", answer)

	return nil
}
