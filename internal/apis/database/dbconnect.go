package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Images struct {
	Id       uint   `gorm:"primaryKey,unique,not_null,autoIncrement,column:idi"`
	FilePath string `gorm:"unique,not_null"`
	XDim     uint   `gorm:"column:xdim"`
	YDim     uint   `gorm:"column:ydim"`
}

func ConnectDb(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Images{})

	return db, nil
}
