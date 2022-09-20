package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testECService = NewEllipticCurve()

func TestEllipticCurve_GenerateKeys(t *testing.T) {
	t.Run("elliptic curve should generate keys", func(t *testing.T) {
		_, _, err := testECService.GenerateKeys()
		assert.Nil(t, err)
	})
}

func TestEllipticCurve_VerifySignature(t *testing.T) {
	t.Run("elliptic curve should pass", func(t *testing.T) {
		pk, pubk, err := testECService.GenerateKeys()
		assert.Nil(t, err)

		encodedPub, err := testECService.EncodePublic(pubk)
		assert.Nil(t, err)
		encodedPriv, err := testECService.EncodePrivate(pk)
		assert.Nil(t, err)

		decodedPub, err := testECService.DecodePublic(encodedPub)
		assert.Nil(t, err)
		decodePriv, err := testECService.DecodePrivate(encodedPriv)
		assert.Nil(t, err)

		signature, err := testECService.CreateSignature([]byte("hello world"), decodePriv)
		assert.Nil(t, err)

		valid, err := testECService.VerifySignature([]byte("hello world"), signature, decodedPub)
		assert.Nil(t, err)
		assert.True(t, valid)
	})

}
