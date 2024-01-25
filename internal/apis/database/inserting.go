package database

import (
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InsertRecords(dbPath string, filesPath []string, logger *zap.Logger) error {
	logger.Debug("Connecting to database", zap.String("dbPath", dbPath))

	conn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	logger.Debug("Iterating through array", zap.Strings("filesPath", filesPath))
	for _, file := range filesPath {
		var (
			image Images
		)

		err = conn.Where("file_path = ?", file).First(&image).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) {

			logger.Debug("File exists, skipping...",
				zap.String("file", file),
			)

			continue
		}

		logger.Debug("Getting the bounds of image",
			zap.String("file", file),
		)

		image.FilePath = file
		x, y, err := GetBounds(image.FilePath)
		if err != nil {
			return err
		}

		image.XDim = x
		image.YDim = y

		logger.Debug("Adding image into database",
			zap.Any("image", image),
			zap.String("dbPath", dbPath),
		)

		rowsAff := conn.FirstOrCreate(&image, "file_path = ?", file).RowsAffected

		logger.Debug("Rows affected",
			zap.Int64("rowsAff", rowsAff),
		)
	}

	return nil
}

func GetBounds(path string) (uint, uint, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0765)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	m, _, err := image.Decode(f)
	if err != nil {
		return 0, 0, err
	}

	return uint(m.Bounds().Dx()), uint(m.Bounds().Dy()), nil
}
