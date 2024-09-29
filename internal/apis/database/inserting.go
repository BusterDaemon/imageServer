package database

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EmptyTagList struct{}

func (e EmptyTagList) Error() string {
	return "Empty tag list"
}

func InsertClientReqRecord(db *gorm.DB, data ClientReqs, logger *zap.Logger) error {
	logger.Debug(
		"Trying to insert log into database",
		zap.Any("clientData", data),
	)

	err := db.Create(data).Error
	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)
		return err
	}

	return nil
}
