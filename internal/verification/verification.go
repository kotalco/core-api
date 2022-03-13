package verification

type Verification struct {
	ID        string
	UserId    string `gorm:"uniqueIndex"`
	Token     string
	ExpiresAt int64
	Completed bool
}
