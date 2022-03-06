package sendgrid

import (
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/sendgrid/sendgrid-go"
	"sync"
)

var SgClient *sendgrid.Client

var lock = &sync.Mutex{}

func GetClient() *sendgrid.Client {
	lock.Lock()
	defer lock.Unlock()
	if SgClient == nil {
		mailClient := sendgrid.NewSendClient(config.EnvironmentConf["SEND_GRID_API_KEY"])
		SgClient = mailClient
	}
	return SgClient
}
