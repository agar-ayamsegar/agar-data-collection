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
	var (
		kind = d.app.Cfg.GetString("database.type")
		user = d.app.Cfg.GetString("database.user")
		pass = d.app.Cfg.GetString("database.pass")
		host = d.app.Cfg.GetString("database.host")
		name = d.app.Cfg.GetString("database.name")
	)

	// A local MongoDB has no user or password, and "mongodb://:@host/db"
	// is not a valid URI, so leave the credentials out entirely.
	if user == "" {
		return fmt.Sprintf("%s://%s/%s", kind, host, name)
	}

	return fmt.Sprintf("%s://%s:%s@%s/%s", kind, user, pass, host, name)
}
