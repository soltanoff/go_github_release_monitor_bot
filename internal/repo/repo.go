package repo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/soltanoff/go_github_release_monitor_bot/internal/config"
	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository() (*Repository, error) {
	// gorm.Config{}: логгировать ошибку на самом высоком уровне и/или использовать свой logger?
	db, err := gorm.Open(sqlite.Open(config.DBName), &gorm.Config{})
	if err != nil {
		slog.Error("[DB] DB connection error", "error", err.Error())
		return nil, fmt.Errorf("[DB] failed to connect database: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) AutoMigrate() error {
	err := r.db.AutoMigrate(&entities.User{}, &entities.Repository{}, &entities.UserRepository{})
	// check error for panic
	if err != nil {
		slog.Error("[DB] DB migration error", "error", err.Error())
		return fmt.Errorf("[DB] failed to migrate database: %w", err)
	}

	return nil
}

func (r *Repository) GetOrCreateUser(
	ctx context.Context,
	userExternalID int64,
) (user entities.User, err error) {
	tx := r.db.Begin().WithContext(ctx)

	defer tx.Rollback()

	if err := tx.First(&user, "external_id = ?", userExternalID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = entities.User{ExternalID: userExternalID}
			if err := tx.Create(&user).Error; err != nil {
				slog.Error("[DB] User creation failure", "senderID", userExternalID, "error", err)
				return user, err
			}
		} else {
			slog.Error("[DB] User creation unexpected error", "senderID", userExternalID, "error", err)
			return user, err
		}
	}

	return user, tx.Commit().Error
}

func (r *Repository) GetAllUserSubscriptions(
	ctx context.Context,
	user *entities.User,
) (selectedRepository []entities.Repository, err error) {
	tx := r.db.Begin().WithContext(ctx)

	defer tx.Rollback()

	tx.Joins("JOIN user_repositories ON repositories.id = user_repositories.repository_id").
		Where("user_repositories.user_id = ?", user.ID).
		Find(&selectedRepository)

	return selectedRepository, nil
}

func (r *Repository) AddUserSubscription(
	ctx context.Context,
	user *entities.User,
	receivedMessage string,
) (err error) {
	tx := r.db.Begin().WithContext(ctx)

	defer tx.Rollback()

	for _, repositoryURL := range strings.Fields(receivedMessage) {
		uriMatches := config.GithubPattern.FindStringSubmatch(repositoryURL)
		if len(uriMatches) == 0 {
			slog.Warn("[DB] Repository skipped by check", "url", repositoryURL)
			continue
		}

		var repository entities.Repository

		query := tx.Where("url = ?", repositoryURL)

		if err := query.First(&repository).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				repository = entities.Repository{URL: repositoryURL, ShortName: uriMatches[1]}
				if err := tx.Create(&repository).Error; err != nil {
					slog.Error("[DB] Repository creation failure", "url", repositoryURL, "error", err)
					return err
				}

				slog.Info("[DB] Repository doesn't exist: create new repository URL", "url", repositoryURL)
			} else {
				slog.Error("[DB] Repository check unexpected error", "url", repositoryURL, "error", err)
				return err
			}
		}

		var userRepo entities.UserRepository

		query = tx.Where("user_id = ? AND repository_id = ?", user.ID, repository.ID)

		if err := query.First(&userRepo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				userRepo = entities.UserRepository{UserID: user.ID, RepositoryID: repository.ID}
				if err := tx.Create(&userRepo).Error; err != nil {
					slog.Error(
						"[DB] UserRepository creation failure",
						"user", user.ID,
						"repo", repository.ID,
						"error", err,
					)

					return err
				}

				slog.Info("[DB] Subscribe user", "user", user.ID, "url", repositoryURL)
			} else {
				slog.Error(
					"[DB] UserRepository check unexpected error",
					"user", user.ID,
					"repo", repository.ID,
					"error", err,
				)

				return err
			}
		}
	}

	return tx.Commit().Error
}

func (r *Repository) RemoveUserSubscription(
	ctx context.Context,
	user *entities.User,
	receivedMessage string,
) (err error) {
	tx := r.db.Begin().WithContext(ctx)

	defer tx.Rollback()

	for _, repositoryURL := range strings.Fields(receivedMessage) {
		if !config.GithubPattern.MatchString(repositoryURL) {
			slog.Warn("[DB] Repository skipped by check", "url", repositoryURL)
			continue
		}

		var repository entities.Repository

		query := tx.Where("url = ?", repositoryURL)

		if err := query.First(&repository).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Info("[DB] Repository doesn't exist", "url", repositoryURL)
				continue
			}

			slog.Warn("[DB] Repository unexpected error", "url", repositoryURL)

			return err
		}

		var userRepo entities.UserRepository

		query = tx.Where("user_id = ? AND repository_id = ?", user.ID, repository.ID)

		if query.First(&userRepo).Error == nil {
			if err := tx.Unscoped().Delete(&userRepo).Error; err != nil {
				slog.Error("[DB] Unsubscribe user error", "user", user.ID, "url", repositoryURL, "error", err)
				return err
			}

			slog.Info("[DB] Unsubscribe user", "user", user.ID, "url", repositoryURL)
		}
	}

	return tx.Commit().Error
}

func (r *Repository) RemoveAllUserSubscriptions(
	ctx context.Context,
	user *entities.User,
) (err error) {
	tx := r.db.Begin().WithContext(ctx)

	defer tx.Rollback()

	var userRepos []entities.UserRepository

	query := tx.Where("user_id = ?", user.ID)

	if err := query.Find(&userRepos).Error; err != nil {
		slog.Error("[DB] UserRepository error", "user", user.ID, "error", err)
		return err
	}

	for index := range userRepos {
		userRepo := userRepos[index]
		if err := tx.Unscoped().Delete(&userRepo).Error; err != nil {
			slog.Error("[DB] UserRepository removing failed", "userRepo", userRepo.ID, "error", err)
			return err
		}
	}

	return tx.Commit().Error
}

func (r *Repository) GetAllRepositories(ctx context.Context) (repositories []entities.Repository, err error) {
	tx := r.db.Begin().WithContext(ctx)

	defer tx.Rollback()

	if err := tx.Find(&repositories).Error; err != nil {
		slog.Error("[DB] Get all repositories failed", "error", err)
		return nil, err
	}

	return repositories, nil
}

func (r *Repository) UpdateRepository(
	ctx context.Context,
	repository *entities.Repository,
) (err error) {
	tx := r.db.Begin().WithContext(ctx)

	defer tx.Rollback()

	if err := tx.Save(&repository).Error; err != nil {
		slog.Error("[DB] Repository update failed", "error", err)
		return err
	}

	return tx.Commit().Error
}

func (r *Repository) GetAllSubscribers(
	ctx context.Context,
	repositoryID uint,
) (users []entities.User, err error) {
	tx := r.db.Begin().WithContext(ctx)

	defer tx.Rollback()

	query := tx.Joins("JOIN user_repositories ur on ur.user_id = users.id")
	if err := query.Where("ur.repository_id = ?", repositoryID).Find(&users).Error; err != nil {
		slog.Error("[DB] Get all subscribers failed", "error", err)
		return nil, err
	}

	return users, nil
}
