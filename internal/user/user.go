package user

type User struct {
	ID         string
	Email      string `gorm:"uniqueIndex"`
	IsVerified bool
	Password   string
}
