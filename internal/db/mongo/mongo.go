package mongo

import (
	"context"
	"fmt"

	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type DB struct {
	app *config.AppConfig
	DB  *mongo.Database
}

func New(app *config.AppConfig) *DB {
	db := &DB{
		app: app,
	}

	client, err := mongo.Connect(
		options.Client().
			ApplyURI(db.DSN()).
			SetMaxPoolSize(uint64(db.app.Cfg.GetInt("database.max_open_conn"))).
			SetMaxConnIdleTime(db.app.Cfg.GetDuration("database.max_open_time")).
			SetMinPoolSize(uint64(db.app.Cfg.GetInt("database.min_open_conn"))),
	)

	if err != nil {
		db.app.Logger.WithError(err).Fatal("Failed to connect to database")
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		db.app.Logger.WithError(err).Fatal("Failed to ping database")
	}

	db.DB = client.Database(db.app.Cfg.GetString("database.name"))

	return db
}

func (d *DB) DSN() string {
	return fmt.Sprintf(
		"%s://%s:%s@%s/%s",
		d.app.Cfg.GetString("database.type"),
		d.app.Cfg.GetString("database.user"),
		d.app.Cfg.GetString("database.pass"),
		d.app.Cfg.GetString("database.host"),
		d.app.Cfg.GetString("database.name"),
	)
}
