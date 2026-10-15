package handler

import (
	"net/http"

	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
	"github.com/labstack/echo/v5"
)

type IHealthHandler interface {
	PingGET(c *echo.Context) error
}

type HealthHandler struct {
	app *config.AppConfig
}

func NewHealthHandler(app *config.AppConfig) IHealthHandler {
	return &HealthHandler{
		app: app,
	}
}

func (h *HealthHandler) PingGET(c *echo.Context) error {
	return c.JSON(http.StatusOK, "pong")
}
