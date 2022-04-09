package security

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var hashingService = NewHashing()

func TestHashing_Hash(t *testing.T) {
	t.Run("Hash_Should_Pass", func(t *testing.T) {
		hash, err := hashingService.Hash("123456", 13)
		assert.Nil(t, err)
		assert.NotNil(t, hash)
		fmt.Println(string(hash))
	})
}

func TestHashing_VerifyHash(t *testing.T) {
	hash := "$2a$13$qe3ZXLa5MS2/NXXL5rxAuuFRoJCLh15qSP9Yf2.Eo6qvk70SkmMwy" //hashed password for 123456
	pass := "123456"
	t.Run("Verify_Hash_Should_Pass", func(t *testing.T) {
		err := hashingService.VerifyHash(hash, pass)
		assert.Nil(t, err)
	})
	t.Run("Verify_Hash_Should_Throw_If_Password_Is_Wrong", func(t *testing.T) {
		err := hashingService.VerifyHash(hash, "invalid")
		assert.Error(t, err, "crypto/bcrypt: hashedPassword is not the hash of the given password")
	})

	t.Run("Verify_Hash_Should_Throw_If_Hash_Is_invalid_Format", func(t *testing.T) {
		err := hashingService.VerifyHash("invalid", pass)
		assert.Error(t, err)
	})

}
