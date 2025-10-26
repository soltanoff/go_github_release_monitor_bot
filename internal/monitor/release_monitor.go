package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/soltanoff/go_github_release_monitor_bot/internal/config"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/monitor/github"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
)

type ReleaseMonitor struct {
	cfg          *config.Config
	bc           *controller.BotController
	repository   *repo.Repository
	githubClient *github.Client
}

func NewReleaseMonitor(
	cfg *config.Config,
	bc *controller.BotController,
	repository *repo.Repository,
) *ReleaseMonitor {
	return &ReleaseMonitor{cfg: cfg, bc: bc, repository: repository, githubClient: github.NewClient()}
}

func (rm *ReleaseMonitor) Start(ctx context.Context) {
	slog.Info("[GITHUB-MONITOR] Starting release monitor...")
	rm.runReleaseMonitor(ctx)
	slog.Info("[GITHUB-MONITOR] Close release monitor...")
}

func (rm *ReleaseMonitor) runReleaseMonitor(ctx context.Context) {
	ticker := time.NewTicker(rm.cfg.SurveyPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rm.dataCollector(ctx)
			// we deliberately reset it, since we need to wait for the
			// specified time from the moment the operation is completed
			ticker.Reset(rm.cfg.SurveyPeriod)
		}
	}
}

func (rm *ReleaseMonitor) dataCollector(ctx context.Context) {
	ticker := time.NewTicker(rm.cfg.FetchingStepPeriod)
	defer ticker.Stop()

	slog.Info("[GITHUB-MONITOR] Start repos data collection")

	repositories, err := rm.repository.GetAllRepositories(ctx)
	if err != nil {
		slog.Error("[GITHUB-MONITOR] Repositories selection unexpected error", "error", err)

		return
	}

	for index := range repositories {
		repository := repositories[index]

		select {
		case <-ctx.Done():
			slog.Info("[GITHUB-MONITOR] Close data collector...")

			return
		case <-ticker.C:
			err := rm.checkLastRepositoryTag(ctx, &repository)
			if err != nil {
				slog.Error(
					"[GITHUB-MONITOR] Data collection error",
					"repository", repository.ShortName,
					"error", err,
				)
			}
			// we deliberately reset it, since we need to wait for the
			// specified time from the moment the operation is completed
			ticker.Reset(rm.cfg.FetchingStepPeriod)
		}
	}

	slog.Info("[GITHUB-MONITOR] Repos data collection is completed")
}

func (rm *ReleaseMonitor) checkLastRepositoryTag(
	ctx context.Context,
	repository *entities.Repository,
) error {
	releaseInfo, err := rm.githubClient.GetLatestTagFromReleaseURI(ctx, repository.ShortName)
	if err != nil {
		return fmt.Errorf("[GITHUB-MONITOR] cannot get latest tag for repository: %w", err)
	}

	if releaseInfo.IsZero() {
		releaseInfo, err = rm.githubClient.GetLatestTagFromTagURI(ctx, repository.ShortName)
		if err != nil {
			return fmt.Errorf("[GITHUB-MONITOR] cannot get latest tag for repository: %w", err)
		}
	}

	if repository.LatestTag == releaseInfo.TagName {
		slog.Info("[GITHUB-MONITOR] Tag exists", "repository", repository.ShortName, "tag", releaseInfo.TagName)

		return nil
	}

	repository.LatestTag = releaseInfo.TagName

	if err := rm.repository.UpdateRepository(ctx, repository); err != nil {
		return fmt.Errorf("[GITHUB-MONITOR] failed to repository: %w", err)
	}

	users, err := rm.repository.GetAllSubscribers(ctx, repository.ID)
	if err != nil {
		slog.Info("[GITHUB-MONITOR]Get subscribers failed", "repository", repository.ShortName, "error", err)

		return fmt.Errorf("[GITHUB-MONITOR] get subscribers failed: %w", err)
	}

	var answer strings.Builder

	answer.WriteString("<b>Release tag</b>: ")
	answer.WriteString(releaseInfo.SourceURL)

	// Source: https://core.telegram.org/bots/faq#my-bot-is-hitting-limits-how-do-i-avoid-this
	// the API will not allow more than 30 messages per second or so
	for _, user := range users {
		err := rm.bc.SendMessage(ctx, user.ExternalID, answer.String(), false)
		if err != nil {
			slog.Warn(
				"[GITHUB-MONITOR] Send message failed",
				"repository", repository.ShortName,
				"receiverID", user.ExternalID,
				"error", err,
			)
		} else {
			slog.Info(
				"[GITHUB-MONITOR] Sending data to user",
				"repository", repository.ShortName,
				"receiverID", user.ExternalID,
			)
		}
	}

	return nil
}
