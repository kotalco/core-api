package setting

type Setting struct {
	Key   string `gorm:"uniqueIndex"`
	Value string
}
