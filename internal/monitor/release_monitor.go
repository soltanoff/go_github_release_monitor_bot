package monitor

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/monitor/github"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/config"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/logs"
)

func Start(ctx context.Context, wg *sync.WaitGroup, botController *controller.BotController) {
	logs.LogInfo("Starting release monitor...")

	wg.Add(1)

	go func() {
		defer wg.Done()
		runReleaseMonitor(ctx, botController)
	}()
}

func runReleaseMonitor(ctx context.Context, botController *controller.BotController) {
	ticker := time.NewTicker(config.SurveyPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logs.LogInfo("Close release monitor...")
			return
		case <-ticker.C:
			dataCollector(ctx, botController)
			// we deliberately reset it, since we need to wait for the
			// specified time from the moment the operation is completed
			ticker.Reset(config.SurveyPeriod)
		}
	}
}

func dataCollector(ctx context.Context, botController *controller.BotController) {
	ticker := time.NewTicker(config.FetchingStepPeriod)
	defer ticker.Stop()

	logs.LogInfo("Start repos data collection")

	repositories, err := repo.GetAllRepositories(ctx)
	if err != nil {
		logs.LogError("Repositories selection unexpected error: %s", err)
		return
	}

	httpClient := http.Client{}

	for index := range repositories {
		repository := repositories[index]

		select {
		case <-ctx.Done():
			logs.LogInfo("Close data collector...")
			return
		case <-ticker.C:
			err := checkLastRepositoryTag(ctx, botController, &httpClient, &repository)
			if err != nil {
				logs.LogError("Data collection error caused for %s: %s", repository.ShortName, err)
			}
			// we deliberately reset it, since we need to wait for the
			// specified time from the moment the operation is completed
			ticker.Reset(config.FetchingStepPeriod)
		}
	}

	logs.LogInfo("Repos data collection is completed")
}

func checkLastRepositoryTag(
	ctx context.Context,
	botController *controller.BotController,
	httpClient *http.Client,
	repository *entities.Repository,
) error {
	releaseInfo, err := github.GetLatestTagFromReleaseURI(ctx, httpClient, repository.ShortName)
	if err != nil {
		return fmt.Errorf("cannot get latest tag for repo: %w", err)
	}

	if releaseInfo.IsZero() {
		releaseInfo, err = github.GetLatestTagFromTagURI(ctx, httpClient, repository.ShortName)
		if err != nil {
			return fmt.Errorf("cannot get latest tag for repo: %w", err)
		}
	}

	if repository.LatestTag == releaseInfo.TagName {
		logs.LogInfo("[%s] Tag %s exists", repository.ShortName, releaseInfo.TagName)
		return nil
	}

	repository.LatestTag = releaseInfo.TagName

	if err := repo.UpdateRepository(ctx, repository); err != nil {
		return fmt.Errorf("failed to repository: %w", err)
	}

	users, err := repo.GetAllSubscribers(ctx, repository.ID)
	if err != nil {
		logs.LogInfo("[%s] Get subscribers failed: %s", repository.ShortName, err)
		return fmt.Errorf("get subscribers failed: %w", err)
	}

	var answer strings.Builder

	answer.WriteString("<b>Release tag</b>: ")
	answer.WriteString(releaseInfo.SourceURL)

	// Source: https://core.telegram.org/bots/faq#my-bot-is-hitting-limits-how-do-i-avoid-this
	// the API will not allow more than 30 messages per second or so
	for _, user := range users {
		err := botController.SendMessage(ctx, user.ExternalID, answer.String(), false)
		if err != nil {
			logs.LogWarn("[%s] Sending to %d has error %s", repository.ShortName, user.ExternalID, err)
		} else {
			logs.LogInfo("[%s] Sending to %d", repository.ShortName, user.ExternalID)
		}
	}

	return nil
}
