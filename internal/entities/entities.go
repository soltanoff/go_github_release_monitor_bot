package entities

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ExternalID int64 `gorm:"not null"`
}

type Repository struct {
	gorm.Model
	LatestTag string `gorm:"size:50"`
	ShortName string `gorm:"size:50;not null;unique"`
	URL       string `gorm:"size:100;not null;unique"`
}

type UserRepository struct {
	gorm.Model
	UserID       uint
	RepositoryID uint
	User         *User       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Repository   *Repository `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
