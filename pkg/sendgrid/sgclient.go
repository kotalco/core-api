package sendgrid

import (
	"sync"

	"github.com/kotalco/cloud-api/pkg/config"
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
