package dbkeystore

type KeyStore struct {
	Key   string `gorm:"uniqueIndex"`
	Value string
}
