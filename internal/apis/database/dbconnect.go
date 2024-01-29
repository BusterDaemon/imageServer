package database

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Images struct {
	Id       uint   `gorm:"primaryKey,unique,not_null,autoIncrement,column:idi"`
	FilePath string `gorm:"unique,not_null"`
	XDim     uint   `gorm:"column:xdim"`
	YDim     uint   `gorm:"column:ydim"`
}

type ClientReqs struct {
	Time    time.Time `gorm:"client_time_access"`
	Ip      string    `gorm:"column:client_ip"`
	Url     string    `gorm:"column:client_url"`
	Queries string    `gorm:"column:client_queries"`
	Ua      string    `gorm:"column:client_ua"`
	Method  string    `gorm:"column:client_method"`
}

func ConnectDb(dbPath string, logger *zap.Logger) (*gorm.DB, error) {
	logger.Debug(
		"Connecting to database",
		zap.String("dbPath", dbPath),
	)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		logger.Error(
			"Can't connect to database",
			zap.Error(err),
		)
		return nil, err
	}
	db.AutoMigrate(
		&Images{},
		&ClientReqs{},
	)

	return db, nil
}
