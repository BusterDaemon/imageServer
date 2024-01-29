package database

import (
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DeleteOptions struct {
	Id       uint
	FilePath string
	UseAnd   bool
}

func DeleteRecord(db *gorm.DB, delOptions DeleteOptions, logger *zap.Logger) error {
	res := db.Model(&Images{})

	logger.Debug(
		"Trying to delete record from database with parameters",
		zap.Any("delOptions", delOptions),
	)

	if delOptions.Id != 0 && delOptions.FilePath != "" {
		switch delOptions.UseAnd {
		case true:
			res.Delete(
				&Images{},
				"id = ? AND file_path = ?",
				delOptions.Id,
				delOptions.FilePath,
			)
		case false:
			res.Delete(
				&Images{},
				"id = ? OR file_path = ?",
				delOptions.Id,
				delOptions.FilePath,
			)
		}
	} else if delOptions.Id != 0 && delOptions.FilePath == "" {
		res.Delete(
			&Images{},
			"id = ?",
			delOptions.Id,
		)
	} else if delOptions.Id == 0 && delOptions.FilePath != "" {
		res.Delete(
			&Images{},
			"file_path = ?",
			delOptions.FilePath,
		)
	} else if delOptions.Id == 0 && delOptions.FilePath == "" {
		logger.Error("Parameters are empty")
		return errors.New("parameters for deletion can't be empty")
	}

	return nil
}
