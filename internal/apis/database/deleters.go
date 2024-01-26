package database

import (
	"errors"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DeleteOptions struct {
	Id       uint
	FilePath string
	UseAnd   bool
}

func DeleteRecord(dbPath string, delOptions DeleteOptions, logger *zap.Logger) error {
	var (
		conn *gorm.DB
		err  error
	)

	logger.Debug(
		"Connecting to database",
		zap.String("dbPath", dbPath),
	)

	if conn, err = gorm.Open(sqlite.Open(dbPath)); err != nil {
		logger.Error(
			"Can't connect to database",
			zap.Error(err),
		)
		return err
	}

	res := conn.Model(&Images{})

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
