package main

import (
	"flag"
	"fmt"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/seeds"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
)

func main() {
	dbClient := sqlclient.OpenDBConnection()
	action := flag.String("a", "", "what to do seed or trunc")
	flag.Parse()
	if config.EnvironmentConf["ENVIRONMENT"] == "development" {
		switch *action {
		case "seed":
			fmt.Println("Seeding...")
			for _, seed := range seeds.All() {
				if err := seed.Run(dbClient); err != nil {
					logger.Error(fmt.Sprintf("SEEDING_ERROR %s", seed.Name), err)
				}
			}
			break
		case "trunc":
			fmt.Println("Truncating...")
			seeds.ClearDB(dbClient)
			break
		default:
			logger.Info("Invalid Command Line")
		}
	} else {
		logger.Info("Seeds only runs on Development environment")
	}

}
