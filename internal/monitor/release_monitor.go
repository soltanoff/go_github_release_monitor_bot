package monitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/monitor/github"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/config"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/logs"
)

type ReleaseMonitor struct {
	bc           *controller.BotController
	repo         *repo.Repository
	githubClient *github.Client
}

func New(
	bc *controller.BotController,
	repo *repo.Repository,
) *ReleaseMonitor {
	return &ReleaseMonitor{bc: bc, repo: repo, githubClient: github.New()}
}

func (rm *ReleaseMonitor) Start(ctx context.Context) {
	logs.LogInfo("[GITHUB-MONITOR] Starting release monitor...")
	rm.runReleaseMonitor(ctx)
	logs.LogInfo("[GITHUB-MONITOR] Close release monitor...")
}

func (rm *ReleaseMonitor) runReleaseMonitor(ctx context.Context) {
	ticker := time.NewTicker(config.SurveyPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rm.dataCollector(ctx)
			// we deliberately reset it, since we need to wait for the
			// specified time from the moment the operation is completed
			ticker.Reset(config.SurveyPeriod)
		}
	}
}

func (rm *ReleaseMonitor) dataCollector(ctx context.Context) {
	ticker := time.NewTicker(config.FetchingStepPeriod)
	defer ticker.Stop()

	logs.LogInfo("[GITHUB-MONITOR] Start repos data collection")

	repositories, err := rm.repo.GetAllRepositories(ctx)
	if err != nil {
		logs.LogError("[GITHUB-MONITOR] Repositories selection unexpected error: %s", err)
		return
	}

	for index := range repositories {
		repository := repositories[index]

		select {
		case <-ctx.Done():
			logs.LogInfo("[GITHUB-MONITOR] Close data collector...")
			return
		case <-ticker.C:
			err := rm.checkLastRepositoryTag(ctx, &repository)
			if err != nil {
				logs.LogError("[GITHUB-MONITOR] Data collection error caused for %s: %s", repository.ShortName, err)
			}
			// we deliberately reset it, since we need to wait for the
			// specified time from the moment the operation is completed
			ticker.Reset(config.FetchingStepPeriod)
		}
	}

	logs.LogInfo("[GITHUB-MONITOR] Repos data collection is completed")
}

func (rm *ReleaseMonitor) checkLastRepositoryTag(
	ctx context.Context,
	repository *entities.Repository,
) error {
	releaseInfo, err := rm.githubClient.GetLatestTagFromReleaseURI(ctx, repository.ShortName)
	if err != nil {
		return fmt.Errorf("[GITHUB-MONITOR] cannot get latest tag for repo: %w", err)
	}

	if releaseInfo.IsZero() {
		releaseInfo, err = rm.githubClient.GetLatestTagFromTagURI(ctx, repository.ShortName)
		if err != nil {
			return fmt.Errorf("[GITHUB-MONITOR] cannot get latest tag for repo: %w", err)
		}
	}

	if repository.LatestTag == releaseInfo.TagName {
		logs.LogInfo("[GITHUB-MONITOR][%s] Tag %s exists", repository.ShortName, releaseInfo.TagName)
		return nil
	}

	repository.LatestTag = releaseInfo.TagName

	if err := rm.repo.UpdateRepository(ctx, repository); err != nil {
		return fmt.Errorf("[GITHUB-MONITOR] failed to repository: %w", err)
	}

	users, err := rm.repo.GetAllSubscribers(ctx, repository.ID)
	if err != nil {
		logs.LogInfo("[GITHUB-MONITOR][%s] Get subscribers failed: %s", repository.ShortName, err)
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
			logs.LogWarn("[GITHUB-MONITOR][%s] Sending to %d has error %s", repository.ShortName, user.ExternalID, err)
		} else {
			logs.LogInfo("[GITHUB-MONITOR][%s] Sending to %d", repository.ShortName, user.ExternalID)
		}
	}

	return nil
}
