package rest

import (
	"fmt"

	"github.com/agar-ayamsegar/agar-data-collection/internal/delivery/rest/handler"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type RestDelivery struct {
	app    *config.AppConfig
	router *echo.Echo

	// handler
	healthHandler handler.IHealthHandler
	gasHandler    handler.IGasHandler
}

func New(app *config.AppConfig) *RestDelivery {
	r := &RestDelivery{
		healthHandler: handler.NewHealthHandler(app),
		gasHandler:    handler.NewGasHandler(app),
		app:           app,
	}

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	r.router = e
	r.RegisterRoutes()
	return r
}

func (r *RestDelivery) Start() error {
	addr := fmt.Sprintf(
		"%s:%d",
		r.app.Cfg.GetString("server.host"),
		r.app.Cfg.GetInt("server.port"),
	)

	return r.router.Start(addr)
}

func (r *RestDelivery) RegisterRoutes() {
	versionRoute := r.router.Group("/v1")

	healthRoute := versionRoute.Group("/health")
	healthRoute.GET("/ping", r.healthHandler.PingGET)
}
