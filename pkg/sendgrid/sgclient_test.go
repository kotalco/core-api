package sendgrid

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetClient(t *testing.T) {
	client := GetClient()
	assert.NotNil(t, client)
}
