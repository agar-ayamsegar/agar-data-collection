package handler

import (
	"errors"
	"net/http"

	"github.com/agar-ayamsegar/agar-data-collection/internal/domain"
	"github.com/agar-ayamsegar/agar-data-collection/internal/usecase"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type IGasHandler interface {
	GasSinglePOST(c *echo.Context) error
	GasImportUpload(c *echo.Context) error
	ListGasByAuthor(c *echo.Context) error
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

func (g *GasHandler) ListGasByAuthor(c *echo.Context) error {
	ctx := c.Request().Context()
	log := g.app.Logger.WithField("method", "ListGasByAuthor")
	trace := c.Response().Header().Get(echo.HeaderXRequestID)

	log.Info("gas request running")
	author := c.Param("author")

	res, err := g.gasUsecase.ListGas(ctx, author)
	if err != nil {
		return ResponseError(c, domain.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
			TraceID: trace,
		}, g.app.Logger)
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse[[]domain.Gas]{
		Data:    res,
		Message: "successfully list gas",
	})
}

func (g *GasHandler) GasImportUpload(c *echo.Context) error {
	ctx := c.Request().Context()
	log := g.app.Logger.WithField("method", "GasImportUpload")
	trace := c.Response().Header().Get(echo.HeaderXRequestID)

	log.Info("gas csv upload request running")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return ResponseError(c, domain.ErrorResponse{
			Message: `no file found in form field "file"`,
			Status:  http.StatusBadRequest,
			TraceID: trace,
		}, g.app.Logger)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return ResponseError(c, domain.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
			TraceID: trace,
		}, g.app.Logger)
	}

	defer src.Close()

	log.WithField("filename", fileHeader.Filename).Info("reading uploaded csv")

	inserted, err := g.gasUsecase.ImportCSV(ctx, src)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrInvalidFile) {
			status = http.StatusBadRequest
		}

		return ResponseError(c, domain.ErrorResponse{
			Message: err.Error(),
			Status:  status,
			TraceID: trace,
		}, g.app.Logger)
	}
	return c.JSON(http.StatusOK, domain.SuccessResponse[int]{
		Message: "successfully create gas",
		Data:    inserted,
	})
	return nil
}
