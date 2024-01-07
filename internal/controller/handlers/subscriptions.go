package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/logs"
)

var errorMessage = "Some error caused, please try latter :("

func MySubscriptionsHandler(ctx context.Context, update *models.Update, user *entities.User) string {
	selectedRepository, err := repo.GetAllUserSubscriptions(ctx, user)
	if err != nil {
		logs.LogBotErrorMessage(update, err)
		return errorMessage
	}

	var answer strings.Builder

	if len(selectedRepository) == 0 {
		answer.WriteString("empty")
	} else {
		for _, repository := range selectedRepository {
			latestTag := "fetch in progress"
			if repository.LatestTag != "" {
				latestTag = repository.LatestTag
			}
			answer.WriteString(fmt.Sprintf("\n%s - %s", latestTag, repository.URL))
		}
	}

	return fmt.Sprintf("Subscriptions: %s", answer.String())
}

func SubscribeHandler(ctx context.Context, update *models.Update, user *entities.User) string {
	err := repo.AddUserSubscription(ctx, user, update.Message.Text)
	if err != nil {
		logs.LogBotErrorMessage(update, err)
		return errorMessage
	}

	return "Successfully subscribed!"
}

func UnsubscribeHandler(ctx context.Context, update *models.Update, user *entities.User) string {
	err := repo.RemoveUserSubscription(ctx, user, update.Message.Text)
	if err != nil {
		logs.LogBotErrorMessage(update, err)
		return errorMessage
	}

	return "Successfully unsubscribed!"
}

func RemoveAllSubscriptionsHandler(ctx context.Context, update *models.Update, user *entities.User) string {
	err := repo.RemoveAllUserSubscriptions(ctx, user)
	if err != nil {
		logs.LogBotErrorMessage(update, err)
		return errorMessage
	}

	return "Successfully unsubscribed!"
}
