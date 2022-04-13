package tfa

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var tfaTestService = NewTfa()

func TestTfa_CreateQRCode_Should_Pass(t *testing.T) {
	buffer, secret, err := tfaTestService.CreateQRCode("testing")
	assert.Nil(t, err)
	assert.NotNil(t, secret)
	assert.NotNil(t, buffer)
}

func TestTfa_CheckOtp(t *testing.T) {
	result := tfaTestService.CheckOtp("123456", "12345")
	assert.False(t, result)
}
