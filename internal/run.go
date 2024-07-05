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

	repository, err := repo.New()
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	err = repository.AutoMigrate()
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	bc, err := controller.New(config.TelegramAPIKey, repository)
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	g.Go(func() error {
		bc.Start(ctx)
		return nil
	})

	releaseMonitor := monitor.New(bc, repository)

	g.Go(func() error {
		releaseMonitor.Start(ctx)
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("[RUNNER] Exited with error", "err", err)
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	return nil
}
