package main

import (
	"github.com/agar-ayamsegar/agar-data-collection/internal/db/mongo"
	"github.com/agar-ayamsegar/agar-data-collection/internal/delivery/rest"
	mongoRepo "github.com/agar-ayamsegar/agar-data-collection/internal/repository/mongo"
	"github.com/agar-ayamsegar/agar-data-collection/internal/usecase"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
)

func main() {
	appCfg := config.New()
	db := mongo.New(appCfg)

	// repo
	gasRepo := mongoRepo.NewGasRepository(db, appCfg)

	// usecase
	gasUsecase := usecase.NewGasUsecase(appCfg, gasRepo)

	// rest delivery
	restDelivery := rest.New(appCfg, gasUsecase)
	if err := restDelivery.Start(); err != nil {
		appCfg.Logger.WithError(err).Fatal("Failed to start server..")
	}
}
