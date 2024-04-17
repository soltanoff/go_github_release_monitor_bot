package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/monitor"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/config"
	"golang.org/x/sync/errgroup"
)

func Run() error {
	g, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	defer cancel()

	config.LoadConfigFromEnv()

	db, err := repo.NewDBConnection()
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	err = repo.AutoMigrate(db)
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	ctx = context.WithValue(ctx, config.DBContextKey, db.WithContext(ctx))

	bc, err := controller.New(config.TelegramAPIKey)
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	g.Go(func() error {
		bc.Start(ctx)
		return nil
	})

	g.Go(func() error {
		monitor.Start(ctx, bc)
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("[RUNNER] Exited with error", "err", err)
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	return nil
}
