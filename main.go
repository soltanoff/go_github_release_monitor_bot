package main

import (
	"context"
	"github.com/soltanoff/go_github_release_monitor_bot/bot_controller"
	dbModels "github.com/soltanoff/go_github_release_monitor_bot/models"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	dbModels.AutoMigrate()

	bc := bot_controller.New(os.Getenv("TELEGRAM_API_KEY"))
	bc.Start(ctx)
}
