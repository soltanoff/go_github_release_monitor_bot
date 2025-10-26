package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/soltanoff/go_github_release_monitor_bot/internal/config"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/monitor"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
	"golang.org/x/sync/errgroup"
)

func Run() error {
	g, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	defer cancel()

	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	repository, err := repo.NewRepository(cfg)
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	err = repository.AutoMigrate()
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	bc, err := controller.NewBotController(cfg, repository)
	if err != nil {
		return fmt.Errorf("[RUNNER]: %w", err)
	}

	g.Go(func() error {
		bc.Start(ctx)

		return nil
	})

	releaseMonitor := monitor.NewReleaseMonitor(cfg, bc, repository)

	g.Go(func() error {
		releaseMonitor.Start(ctx)

		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("[RUNNER] Exited with error", "error", err)

		return fmt.Errorf("[RUNNER]: %w", err)
	}

	return nil
}
