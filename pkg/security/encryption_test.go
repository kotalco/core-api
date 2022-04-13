package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var encryptionTestService = NewEncryption()

func Test_createHash(t *testing.T) {
	hash := createHash("test")
	assert.EqualValues(t, "098f6bcd4621d373cade4e832627b4f6", hash)
}

func TestEncryption_Encrypt(t *testing.T) {
	encodedCipher, err := encryptionTestService.Encrypt([]byte("test"), "passphrase")
	assert.Nil(t, err)
	assert.NotNil(t, encodedCipher)
}

func TestEncryption_Decrypt(t *testing.T) {
	cipher := "YXY7YYC3ORJC4WSOAEYYGPLRHXQOO3UPMXCG5E44MOCGATNCXNRA====" //this is cipher for the word test
	passPhrase := "passphrase"                                           //pass phrase that we use as a secret for encryption and decryption
	t.Run("Decrypt_Should_Throw_If_Passed_String_Is_Not_Base_32", func(t *testing.T) {
		str, err := encryptionTestService.Decrypt("string", passPhrase)
		assert.EqualValues(t, "", str)
		assert.EqualValues(t, "illegal base32 data at input byte 0", err.Error())

	})

	t.Run("Decrypt_Should_Throw_If_Passed_Pass_Phrase_Is_Different_Than_The_Encrypt_Pass_Phrase", func(t *testing.T) {
		str, err := encryptionTestService.Decrypt(cipher, "invalid")
		assert.EqualValues(t, "cipher: message authentication failed", err.Error())
		assert.EqualValues(t, "", str)
	})

	t.Run("Decrypt_Should_Pass", func(t *testing.T) {
		str, err := encryptionTestService.Decrypt(cipher, passPhrase)
		assert.EqualValues(t, "test", str)
		assert.Nil(t, err)
	})
}
