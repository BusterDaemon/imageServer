package database

import (
	"buster_daemon/imageserver/internal/config"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Images struct {
	Id           uint      `gorm:"primaryKey,unique,not_null,autoIncrement"`
	FilePath     string    `gorm:"unique,not_null"`
	XDim         uint      `gorm:"column:xdim"`
	YDim         uint      `gorm:"column:ydim"`
	DateAdded    time.Time `gorm:"column:added_at"`
	DateCreated  time.Time `gorm:"column:created_at"`
	DateModified time.Time `gorm:"column:modified_at"`
}

type ClientReqs struct {
	Time       time.Time `gorm:"client_time_access"`
	Ip         string    `gorm:"column:client_ip"`
	Url        string    `gorm:"column:client_url"`
	Queries    string    `gorm:"column:client_queries"`
	Ua         string    `gorm:"column:client_ua"`
	Method     string    `gorm:"column:client_method"`
	StatusCode int       `gorm:"column:client_status"`
}

func ConnectDb(config config.Config, logger *zap.Logger) (*gorm.DB, error) {
	var (
		dbDriver gorm.Dialector
	)
	logger.Debug(
		"Connecting to database",
		zap.Any("dbPath", config.Database),
	)

	switch config.Database.DbType {
	case "sqlite":
		dbDriver = sqlite.Open(config.Database.DbAddress)
	case "mysql":
		dbDriver = mysql.New(
			mysql.Config{
				DSN: fmt.Sprintf(
					"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
					config.Database.DbLogin,
					config.Database.DbPassword,
					config.Database.DbAddress,
					config.Database.DbPort,
					config.Database.DbName,
				),
				DefaultStringSize: 256,
			},
		)
	case "postgres":
		dbDriver = postgres.New(
			postgres.Config{
				DSN: fmt.Sprintf(
					"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
					config.Database.DbAddress,
					config.Database.DbLogin,
					config.Database.DbPassword,
					config.Database.DbName,
					config.Database.DbPort,
					config.Database.DbSSL,
				),
			},
		)
	}

	db, err := gorm.Open(dbDriver, &gorm.Config{})
	if err != nil {
		logger.Error(
			"Can't connect to database",
			zap.Error(err),
		)
		return nil, err
	}
	err = db.AutoMigrate(
		&Images{},
		&ClientReqs{},
	)
	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)
		return nil, err
	}

	return db, nil
}
