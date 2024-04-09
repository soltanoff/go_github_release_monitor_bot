package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/monitor"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/config"
)

func main() {
	wg := sync.WaitGroup{}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	defer cancel()

	config.LoadConfigFromEnv()

	db := repo.InitDBConnection()
	repo.AutoMigrate(db)

	ctx = context.WithValue(ctx, config.DBContextKey, db.WithContext(ctx))

	bc := controller.New(config.TelegramAPIKey)
	bc.Start(ctx, &wg)

	monitor.Start(ctx, &wg, &bc)

	wg.Wait()
}
