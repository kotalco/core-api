package user

type User struct {
	ID               string
	Email            string `gorm:"uniqueIndex"`
	IsEmailVerified  bool
	Password         string
	TwoFactorCipher  string
	TwoFactorEnabled bool
	PlatformAdmin    bool `gorm:"default:false"`
}
