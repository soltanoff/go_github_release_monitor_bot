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
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/logs"
)

type BotController struct {
	bot         bot.Bot
	commandList []string
}

type HandlerFunc func(ctx context.Context, update *models.Update, user *entities.User) string

func New(telegramAPIKey string) BotController {
	b, err := bot.New(telegramAPIKey)
	if err != nil {
		logs.LogError("Bot init error: %s", err.Error())
		panic(err)
	}

	bc := BotController{bot: *b}
	bc.registerDefaultHandler()
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
		ChatID:                userExternalID,
		Text:                  answer,
		ParseMode:             models.ParseModeHTML,
		DisableWebPagePreview: disableWebPagePreview,
	})

	if err != nil {
		return fmt.Errorf("send message failed: %w", err)
	}

	logs.LogInfo("<<< User %d: %s", userExternalID, answer)

	return nil
}

func (bc *BotController) getCommandList() string {
	return strings.Join(bc.commandList, "\n")
}

func (bc *BotController) registerDefaultHandler() {
	bot.WithDefaultHandler(bc.handlerWrapper(bc.defaultHandler, true))(&bc.bot)
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

	bc.commandList = append(bc.commandList, fmt.Sprintf("%s - %s", pattern, description))
}

func (bc *BotController) handlerWrapper(handler HandlerFunc, disableWebPagePreview bool) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		logs.LogBotIncommingMessage(update)

		_, err := bc.bot.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: update.Message.Chat.ID,
			Action: models.ChatActionTyping,
		})
		if err != nil {
			logs.LogBotErrorMessage(update, err)
			return
		}

		user, err := repo.GetOrCreateUser(ctx, update.Message.From.ID)
		if err != nil {
			logs.LogBotErrorMessage(update, err)
			return
		}

		answer := handler(ctx, update, &user)

		_, err = bc.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:                update.Message.Chat.ID,
			Text:                  answer,
			ParseMode:             models.ParseModeHTML,
			ReplyToMessageID:      update.Message.ID,
			DisableWebPagePreview: disableWebPagePreview,
		})
		if err != nil {
			logs.LogBotErrorMessage(update, err)
			return
		}

		logs.LogBotOutgoingMessage(update, answer)
	}
}

func (bc *BotController) defaultHandler(_ context.Context, _ *models.Update, _ *entities.User) string {
	return "Say /help"
}

func (bc *BotController) welcomeHandler(_ context.Context, _ *models.Update, _ *entities.User) string {
	return bc.getCommandList()
}
