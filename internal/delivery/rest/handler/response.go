package handler

import (
	"github.com/agar-ayamsegar/agar-data-collection/internal/domain"
	"github.com/labstack/echo/v5"
	"github.com/sirupsen/logrus"
)

func ResponseError(c *echo.Context, err domain.ErrorResponse, log *logrus.Logger) error {
	log.WithField("err", err).Error("error response when calling api")
	return c.JSON(err.Status, err)
}
