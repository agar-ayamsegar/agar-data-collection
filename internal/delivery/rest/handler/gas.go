package handler

import (
	"net/http"

	"github.com/agar-ayamsegar/agar-data-collection/internal/domain"
	"github.com/agar-ayamsegar/agar-data-collection/internal/usecase"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type IGasHandler interface {
	GasSinglePOST(c *echo.Context) error
}

type GasHandler struct {
	app        *config.AppConfig
	gasUsecase usecase.IGasUsecase
}

func NewGasHandler(app *config.AppConfig, gasUsecase usecase.IGasUsecase) IGasHandler {
	return &GasHandler{
		app:        app,
		gasUsecase: gasUsecase,
	}
}

func (g *GasHandler) GasSinglePOST(c *echo.Context) error {
	ctx := c.Request().Context()
	log := g.app.Logger.WithField("method", "GasPOST")
	trace := c.Response().Header().Get(echo.HeaderXRequestID)
	id := uuid.New().String()

	log.Info("gas request running")

	var gas domain.Gas
	if err := c.Bind(&gas); err != nil {
		return ResponseError(c, domain.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
			TraceID: trace,
		}, g.app.Logger)
	}

	log = log.WithField("gas", gas)

	if err := c.Validate(gas); err != nil {
		return ResponseError(c, domain.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
			TraceID: trace,
		}, g.app.Logger)
	}

	gas.ID = id

	// call usecase
	d, err := g.gasUsecase.CreateGas(ctx, gas)
	if err != nil {
		return ResponseError(c, domain.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
			TraceID: trace,
		}, g.app.Logger)
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse[domain.Gas]{
		Message: "successfully create gas",
		Data:    d,
	})
}
