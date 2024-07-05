package handlers

import (
	"context"
	"strings"

	"github.com/go-telegram/bot/models"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/controller/logs"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/repo"
)

const (
	errorMessage               string = "Some error caused, please try latter :("
	emptySubscriptions         string = "Subscriptions: empty"
	emptyString                string = ""
	subscriptionHeader         string = "Subscriptions: "
	fallbackTag                string = "fetch in progress"
	newLineTag                 string = "\n"
	delim                      string = " - "
	successSubscribedMessage   string = "Successfully subscribed!"
	successUnsubscribedMessage string = "Successfully unsubscribed!"
)

type SubscriptionsHandler struct {
	repo *repo.Repository
}

func New(repo *repo.Repository) *SubscriptionsHandler {
	return &SubscriptionsHandler{repo: repo}
}

func (h *SubscriptionsHandler) MySubscriptionsHandler(
	ctx context.Context,
	update *models.Update,
	user *entities.User,
) string {
	selectedRepository, err := h.repo.GetAllUserSubscriptions(ctx, user)
	if err != nil {
		logs.LogBotErrorMessage(update, err)
		return errorMessage
	}

	var answer strings.Builder

	if len(selectedRepository) == 0 {
		return emptySubscriptions
	}

	answer.WriteString(subscriptionHeader)

	for _, repository := range selectedRepository {
		latestTag := fallbackTag
		if repository.LatestTag != emptyString {
			latestTag = repository.LatestTag
		}

		answer.WriteString(newLineTag)
		answer.WriteString(latestTag)
		answer.WriteString(delim)
		answer.WriteString(repository.URL)
	}

	return answer.String()
}

func (h *SubscriptionsHandler) SubscribeHandler(
	ctx context.Context,
	update *models.Update,
	user *entities.User,
) string {
	err := h.repo.AddUserSubscription(ctx, user, update.Message.Text)
	if err != nil {
		logs.LogBotErrorMessage(update, err)
		return errorMessage
	}

	return successSubscribedMessage
}

func (h *SubscriptionsHandler) UnsubscribeHandler(
	ctx context.Context,
	update *models.Update,
	user *entities.User,
) string {
	err := h.repo.RemoveUserSubscription(ctx, user, update.Message.Text)
	if err != nil {
		logs.LogBotErrorMessage(update, err)
		return errorMessage
	}

	return successUnsubscribedMessage
}

func (h *SubscriptionsHandler) RemoveAllSubscriptionsHandler(
	ctx context.Context,
	update *models.Update,
	user *entities.User,
) string {
	err := h.repo.RemoveAllUserSubscriptions(ctx, user)
	if err != nil {
		logs.LogBotErrorMessage(update, err)
		return errorMessage
	}

	return successUnsubscribedMessage
}
