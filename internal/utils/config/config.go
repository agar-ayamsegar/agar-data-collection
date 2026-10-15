package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Cfg    *viper.Viper
	Logger *logrus.Logger
}

func New() *AppConfig {
	cfg := viper.New()
	cfg.SetConfigName("config")
	cfg.SetConfigType("json")
	cfg.AddConfigPath("env")

	err := cfg.ReadInConfig()
	if err != nil {
		logrus.WithError(err).Fatal("error reading config")
	}

	level := logrus.Level(cfg.GetInt("log.level"))
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(level)

	return &AppConfig{
		Cfg:    cfg,
		Logger: logger,
	}
}
