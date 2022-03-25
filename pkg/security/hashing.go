package security

import "golang.org/x/crypto/bcrypt"

type hashing struct {}

type IHashing interface {
	Hash(password string, cost int) ([]byte, error)
	VerifyHash(hashedPassword, password string) error

}

func NewHashing() IHashing {
	newHashing := &hashing{}
	return newHashing
}



func (hashing)Hash(password string, cost int) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), cost)
}

func (hashing)VerifyHash(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
