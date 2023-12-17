package bot_controller

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/utils"
	"log/slog"
	"strings"
)

type BotController struct {
	bot         *bot.Bot
	commandList []string
}

func New(telegramAPIKey string) *BotController {
	b, err := bot.New(telegramAPIKey)
	if err != nil {
		slog.Error(fmt.Sprintf("Bot init error: %s", err.Error()))
		panic(err)
	}

	bc := &BotController{bot: b}

	bc.registerHandler("/start", "base command for user registration", bc.welcomeHandler)
	bc.registerHandler("/help", "view all commands", bc.welcomeHandler)
	bot.WithDefaultHandler(bc.defaultHandler)(bc.bot)

	return bc
}

func (bc *BotController) Start(ctx context.Context) {
	slog.Info("Starting bot...")
	bc.bot.Start(ctx)
}

func (bc *BotController) getCommandList() string {
	return strings.Join(bc.commandList, "\n")
}

func (bc *BotController) registerHandler(pattern string, description string, handler bot.HandlerFunc) {
	bc.bot.RegisterHandler(bot.HandlerTypeMessageText, pattern, bot.MatchTypeExact, handler)
	bc.commandList = append(bc.commandList, fmt.Sprintf("%s - %s", pattern, description))
}

func (bc *BotController) sendChatTypingAction(ctx context.Context, update *models.Update) {
	_, err := bc.bot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: update.Message.Chat.ID,
		Action: models.ChatActionTyping,
	})
	if err != nil {
		slog.Error(err.Error())
	}
}

func (bc *BotController) defaultHandler(ctx context.Context, _ *bot.Bot, update *models.Update) {
	utils.LogIncommingMessage(update)
	bc.sendChatTypingAction(ctx, update)

	answer := "Say /hello"

	_, err := bc.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:           update.Message.Chat.ID,
		Text:             answer,
		ParseMode:        models.ParseModeHTML,
		ReplyToMessageID: update.Message.ID,
	})
	if err != nil {
		slog.Error(err.Error())
	}
	utils.LogOutgoingMessage(update, answer)
}

func (bc *BotController) welcomeHandler(ctx context.Context, _ *bot.Bot, update *models.Update) {
	utils.LogIncommingMessage(update)
	bc.sendChatTypingAction(ctx, update)

	answer := bc.getCommandList()

	_, err := bc.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:           update.Message.Chat.ID,
		Text:             answer,
		ParseMode:        models.ParseModeHTML,
		ReplyToMessageID: update.Message.ID,
	})
	if err != nil {
		slog.Error(err.Error())
	}
	utils.LogOutgoingMessage(update, answer)
}
