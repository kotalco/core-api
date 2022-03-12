package user

type User struct {
	ID              string
	Email           string `gorm:"uniqueIndex"`
	IsEmailVerified bool
	Password        string
}
