package database

import (
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
	db.AutoMigrate(&Images{})

	return db, nil
}
