package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/soltanoff/go_github_release_monitor_bot/internal/entities"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/config"
	"github.com/soltanoff/go_github_release_monitor_bot/pkg/logs"
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
		logs.LogError("[DB] DB connection error: %s", err.Error())
		return nil, fmt.Errorf("[DB] failed to connect database: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) AutoMigrate() error {
	err := r.db.AutoMigrate(&entities.User{}, &entities.Repository{}, &entities.UserRepository{})
	// check error for panic
	if err != nil {
		logs.LogError("[DB] DB migration error: %s", err.Error())
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
				logs.LogError("[DB] User `%d` creation failure: %s", userExternalID, err)
				return user, err
			}
		} else {
			logs.LogError("[DB] User `%d` unexpected error: %s", userExternalID, err)
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
			logs.LogWarn("[DB] Repository skipped by check: %s", repositoryURL)
			continue
		}

		var repository entities.Repository

		query := tx.Where("url = ?", repositoryURL)

		if err := query.First(&repository).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				repository = entities.Repository{URL: repositoryURL, ShortName: uriMatches[1]}
				if err := tx.Create(&repository).Error; err != nil {
					logs.LogError("[DB] Repository `%s` creation failure: %s", repositoryURL, err)
					return err
				}

				logs.LogInfo("[DB] Repository `%s` doesn't exist: create new repository URL", repositoryURL)
			} else {
				logs.LogError("[DB] Repository `%s` check unexpected error: %s", repositoryURL, err)
				return err
			}
		}

		var userRepo entities.UserRepository

		query = tx.Where("user_id = ? AND repository_id = ?", user.ID, repository.ID)

		if err := query.First(&userRepo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				userRepo = entities.UserRepository{UserID: user.ID, RepositoryID: repository.ID}
				if err := tx.Create(&userRepo).Error; err != nil {
					logs.LogError("[DB] UserRepository `%d-%d` creation failure: %s", user.ID, repository.ID, err)
					return err
				}

				logs.LogInfo("[DB] Subscribe user %d to %s", user.ID, repositoryURL)
			} else {
				logs.LogError("[DB] UserRepository `%d-%d` check unexpected error: %s", user.ID, repository.ID, err)
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
			logs.LogWarn("[DB] Repository skipped by check: %s", repositoryURL)
			continue
		}

		var repository entities.Repository

		query := tx.Where("url = ?", repositoryURL)

		if err := query.First(&repository).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logs.LogInfo("[DB] Repository `%s` doesn't exist", repositoryURL)
				continue
			}

			logs.LogWarn("[DB] Repository `%s` unexpected error", repositoryURL)

			return err
		}

		var userRepo entities.UserRepository

		query = tx.Where("user_id = ? AND repository_id = ?", user.ID, repository.ID)

		if query.First(&userRepo).Error == nil {
			if err := tx.Unscoped().Delete(&userRepo).Error; err != nil {
				logs.LogError("[DB] Unsubscribe user %d from %s error: %s", user.ID, repositoryURL, err)
				return err
			}

			logs.LogInfo("[DB] Unsubscribe user %d from %s", user.ID, repositoryURL)
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
		logs.LogError("[DB] UserRepository for userID=%d error: %s", user.ID, err)
		return err
	}

	for index := range userRepos {
		userRepo := userRepos[index]
		if err := tx.Unscoped().Delete(&userRepo).Error; err != nil {
			logs.LogError("[DB] UserRepository %d removing failed: %s", userRepo.ID, err)
			return err
		}
	}

	return tx.Commit().Error
}

func (r *Repository) GetAllRepositories(ctx context.Context) (repositories []entities.Repository, err error) {
	tx := r.db.Begin().WithContext(ctx)

	defer tx.Rollback()

	if err := tx.Find(&repositories).Error; err != nil {
		logs.LogError("[DB] Get all repositories failed: %s", err)
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
		logs.LogError("[DB] Repository update failed: %s", err)
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
		logs.LogError("[DB] Get all subscribers failed: %s", err)
		return nil, err
	}

	return users, nil
}
