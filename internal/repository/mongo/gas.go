package mongo

import (
	"context"

	"github.com/agar-ayamsegar/agar-data-collection/internal/db/mongo"
	"github.com/agar-ayamsegar/agar-data-collection/internal/domain"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IGasRepository interface {
	InsertLog(ctx context.Context, gas domain.Gas) error
	ListLogByAuthor(ctx context.Context, author string) ([]domain.Gas, error)
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

func (repo *GasRepository) ListLogByAuthor(ctx context.Context, author string) ([]domain.Gas, error) {
	cur, err := repo.db.DB.Collection("gas").Find(ctx, bson.M{"author": author})
	if err != nil {
		repo.app.Logger.WithError(err).Error(ctx, "ListLogByAuthor: error when finding gas")
		return nil, err
	}

	defer cur.Close(ctx)

	var gas []domain.Gas

	if err := cur.All(ctx, &gas); err != nil {
		repo.app.Logger.WithError(err).Error(ctx, "ListLogByAuthor: error when finding gas")
		return nil, err
	}

	return gas, nil
}
