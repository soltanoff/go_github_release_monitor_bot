package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller/handlers"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/logs"
)

type BotController struct {
	bot         *bot.Bot
	commandList []string
}

type HandlerFunc func(ctx context.Context, update *models.Update, user *entities.User) string

func New(telegramAPIKey string) BotController {
	// bot.WithErrorsHandler(): логгировать ошибку на самом высоком уровне и/или использовать свой logger?
	b, err := bot.New(telegramAPIKey)
	if err != nil {
		logs.LogError("Bot init error: %s", err.Error())
		panic(err)
	}

	bc := BotController{bot: b}
	bc.registerDefaultMiddlewares()
	bc.registerDefaultHandler()
	bc.registerHandler(
		"/my_subscriptions",
		"view all subscriptions",
		true,
		handlers.MySubscriptionsHandler,
	)
	bc.registerHandler(
		"/subscribe",
		"[github repo urls] subscribe to the new GitHub repository",
		true,
		handlers.SubscribeHandler,
	)
	bc.registerHandler(
		"/unsubscribe",
		"[github repo urls] unsubscribe from the GitHub repository",
		true,
		handlers.UnsubscribeHandler,
	)
	bc.registerHandler(
		"/remove_all_subscriptions",
		"remove all exists subscriptions",
		true,
		handlers.RemoveAllSubscriptionsHandler,
	)

	return bc
}

func (bc *BotController) Start(ctx context.Context, wg *sync.WaitGroup) {
	logs.LogInfo("Starting bot...")

	wg.Add(1)

	go func() {
		defer wg.Done()
		bc.bot.Start(ctx)
		logs.LogInfo("Close bot controller...")
	}()
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
		return fmt.Errorf("send message failed: %w", err)
	}

	logs.LogInfo("<<< User %d: %s", userExternalID, answer)

	return nil
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
