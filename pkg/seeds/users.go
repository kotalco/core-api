package seeds

import (
	"github.com/kotalco/cloud-api/internal/user"
	"gorm.io/gorm"
)

func CreateUser(db *gorm.DB, user *user.User) error {
	return db.Create(user).Error
}
