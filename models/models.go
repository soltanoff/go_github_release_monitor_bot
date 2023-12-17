package models

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log/slog"
)

var DbName = "db.sqlite3"

type User struct {
	gorm.Model
	ExternalID uint
}

type Repository struct {
	gorm.Model
	Url       string
	LatestTag string
}

type UserRepository struct {
	gorm.Model
	UserID       uint
	RepositoryID uint

	User       User
	Repository Repository
}

func AutoMigrate() {
	db, err := gorm.Open(sqlite.Open(DbName), &gorm.Config{})
	if err != nil {
		slog.Error(fmt.Sprintf("DB connection error: %s", err.Error()))
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&User{}, &Repository{}, &UserRepository{})
	if err != nil {
		slog.Error(fmt.Sprintf("DB migration error: %s", err.Error()))
		panic("failed to migrate database schema")
	}
}
