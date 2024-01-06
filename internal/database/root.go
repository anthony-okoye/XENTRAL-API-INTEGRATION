package database

import (
	"bookbox-backend/pkg/logger"
	"os"

	"go.uber.org/zap"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

var database DBConfig

func init() {

	database = DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		Database: os.Getenv("DB_NAME"),
	}

	err := Connect()
	if err != nil {
		logger.Log.Fatal("failed to start database : ",
			zap.Error(err),
		)
		return
	}

	err = Seed()
	if err != nil {
		logger.Log.Error("failed to seed database : ",
			zap.Error(err),
		)
		return
	}
}
