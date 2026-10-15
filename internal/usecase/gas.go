package usecase

import (
	"context"

	"github.com/agar-ayamsegar/agar-data-collection/internal/domain"
	"github.com/agar-ayamsegar/agar-data-collection/internal/repository/mongo"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
)

type IGasUsecase interface {
	CreateGas(ctx context.Context, gas domain.Gas) (domain.Gas, error)
	ListGas(ctx context.Context, author string) ([]domain.Gas, error)
}

type GasUsecase struct {
	app  *config.AppConfig
	repo mongo.IGasRepository
}

func NewGasUsecase(app *config.AppConfig, repo mongo.IGasRepository) IGasUsecase {
	return &GasUsecase{
		app:  app,
		repo: repo,
	}
}

func (g *GasUsecase) CreateGas(ctx context.Context, gas domain.Gas) (domain.Gas, error) {
	log := g.app.Logger.WithField("method", "CreateGas")

	err := g.repo.InsertLog(ctx, gas)
	if err != nil {
		log.WithError(err).Error("failed to call db insert gas")
		return domain.Gas{}, err
	}

	log.WithField("id", gas.ID).Info("creating the gas successfully")

	return gas, nil
}

func (g *GasUsecase) ListGas(ctx context.Context, author string) ([]domain.Gas, error) {
	log := g.app.Logger.WithField("method", "ListGasByAuthor")

	gas, err := g.repo.ListLogByAuthor(ctx, author)
	if err != nil {
		log.WithError(err).Error("failed to call db list gas by author")
		return nil, err
	}

	return gas, nil
}
