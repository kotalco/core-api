package sendgrid

import (
	"github.com/kotalco/core-api/config"
	"sync"

	"github.com/sendgrid/sendgrid-go"
)

var (
	SgClient   *sendgrid.Client
	clientOnce sync.Once
)

func GetClient() *sendgrid.Client {
	clientOnce.Do(func() {
		SgClient = sendgrid.NewSendClient(config.Environment.SendgridAPIKey)
	})
	return SgClient
}
