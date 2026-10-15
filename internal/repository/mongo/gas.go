package mongo

import (
	"context"

	"github.com/agar-ayamsegar/agar-data-collection/internal/db/mongo"
	"github.com/agar-ayamsegar/agar-data-collection/internal/domain"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
)

type IGasRepository interface {
	InsertLog(ctx context.Context, gas domain.Gas) error
}

type GasRepository struct {
	db  *mongo.DB
	app *config.AppConfig
}

func NewGasRepository(db *mongo.DB, app *config.AppConfig) IGasRepository {
	return &GasRepository{
		db:  db,
		app: app,
	}
}

func (repo *GasRepository) InsertLog(ctx context.Context, gas domain.Gas) error {
	_, err := repo.db.DB.Collection("gas").InsertOne(ctx, gas)
	if err != nil {
		repo.app.Logger.WithError(err).Error(ctx, "InsertLog: error when inserting gas")
		return err
	}

	return nil
}
