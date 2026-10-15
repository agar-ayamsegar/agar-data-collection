package handler

import "github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"

type IGasHandler interface {
}

type GasHandler struct {
	app *config.AppConfig
}

func NewGasHandler(app *config.AppConfig) *GasHandler {
	return &GasHandler{
		app: app,
	}
}
