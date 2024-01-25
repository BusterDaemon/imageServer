package database

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DeleteOptions struct {
	Id       uint
	FilePath string
	UseAnd   bool
}

func DeleteRecord(dbPath string, delOptions DeleteOptions) error {
	var (
		conn *gorm.DB
		err  error
	)
	if conn, err = gorm.Open(sqlite.Open(dbPath)); err != nil {
		return err
	}

	res := conn.Model(&Images{})

	if delOptions.Id != 0 && delOptions.FilePath != "" {
		switch delOptions.UseAnd {
		case true:
			res.Delete(&Images{},
				"id = ? AND file_path = ?",
				delOptions.Id,
				delOptions.FilePath,
			)
		case false:
			res.Delete(&Images{},
				"id = ? OR file_path = ?",
				delOptions.Id,
				delOptions.FilePath,
			)
		}
	} else if delOptions.Id != 0 && delOptions.FilePath == "" {
		res.Delete(&Images{},
			"id = ?",
			delOptions.Id,
		)
	} else if delOptions.Id == 0 && delOptions.FilePath != "" {
		res.Delete(&Images{},
			"file_path = ?",
			delOptions.FilePath,
		)
	} else if delOptions.Id == 0 && delOptions.FilePath == "" {
		return errors.New("parameters for deletion can't be empty")
	}

	return nil
}
