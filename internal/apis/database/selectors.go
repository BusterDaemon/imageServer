package database

import (
	"errors"

	"gorm.io/gorm"
)

func GetImageTag(db *gorm.DB, tagName string) (*Tag, error) {
	var (
		tag Tag = Tag{Value: tagName}
	)
	res := db.Where("value = ?", tagName).FirstOrCreate(&tag)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, res.Error
	}

	return &tag, nil
}

func GetUser(db *gorm.DB, username string) (*User, error) {
	var (
		usvr User
	)

	if res := db.Where("login = ?", username).First(&usvr); res.Error != nil {
		return nil, res.Error
	}

	return &usvr, nil
}
