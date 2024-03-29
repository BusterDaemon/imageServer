package main

import (
	"buster_daemon/imageserver/internal/apis"
	"buster_daemon/imageserver/internal/apis/database"
	"buster_daemon/imageserver/internal/config"
	"flag"

	"go.uber.org/zap"
)

func main() {
	var confPath string
	var logger *zap.Logger

	flag.StringVar(&confPath, "config", "./config.json", "Path to config file")
	flag.Parse()

	defer logger.Sync()

	conf := config.New()
	err := (&conf).ReadConfig(confPath)
	if err != nil {
		panic(err)
	}

	switch conf.Logger.LogMode {
	case "dev":
		logger, _ = zap.NewDevelopment()
	case "prod":
		logger, _ = zap.NewProduction()
	default:
		logger, _ = zap.NewProduction()
	}

	logger.Info("Connecting to Database", zap.Any("path", conf.Database))
	db, err := database.ConnectDb(conf, logger)
	if err != nil {
		logger.Fatal("Can't connect to database", zap.Error(err))
	}

	apis.Start(&conf, db, logger)
}
