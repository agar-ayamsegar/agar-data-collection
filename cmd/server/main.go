package main

import (
	"fmt"

	"github.com/agar-ayamsegar/agar-data-collection/internal/db/mongo"
	"github.com/agar-ayamsegar/agar-data-collection/internal/delivery/rest"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
)

func main() {
	appCfg := config.New()
	db := mongo.New(appCfg)

	fmt.Println(db.DSN())

	// rest delivery
	restDelivery := rest.New(appCfg)
	if err := restDelivery.Start(); err != nil {
		appCfg.Logger.WithError(err).Fatal("Failed to start server..")
	}
}
