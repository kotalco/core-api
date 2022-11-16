package dbconfig

type DbConfig struct {
	Key   string `gorm:"uniqueIndex"`
	Value string
}
