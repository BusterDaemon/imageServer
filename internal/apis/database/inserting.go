package database

import (
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"time"

	"github.com/djherbis/times"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type imageData struct {
	xdim         uint
	ydim         uint
	dateCreated  time.Time
	dateModified time.Time
}

func InsertClientReqRecord(dbPath string, data ClientReqs, logger *zap.Logger) error {
	logger.Debug(
		"Connecting to database",
		zap.String("dbPath", dbPath),
	)

	conn, err := gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)
		return err
	}

	logger.Debug(
		"Trying to insert log into database",
		zap.Any("clientData", data),
	)

	err = conn.Create(data).Error
	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)
		return err
	}

	return nil
}

func InsertRecords(dbPath string, filesPath []string, logger *zap.Logger) error {
	logger.Debug(
		"Connecting to database",
		zap.String("dbPath", dbPath),
	)

	conn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	logger.Debug(
		"Iterating through array",
		zap.Strings("filesPath", filesPath),
	)

	for _, file := range filesPath {
		var (
			image Images
		)

		err = conn.Where("file_path = ?", file).First(&image).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) {

			logger.Debug(
				"File exists, skipping...",
				zap.String("file", file),
			)

			continue
		}

		logger.Debug(
			"Getting the bounds of image",
			zap.String("file", file),
		)

		image.FilePath = file
		imageData, err := GetBounds(image.FilePath)
		if err != nil {
			return err
		}

		image.DateAdded = time.Now()
		image.DateCreated = imageData.dateCreated
		image.DateModified = imageData.dateModified
		image.XDim = imageData.xdim
		image.YDim = imageData.ydim

		logger.Debug(
			"Adding image into database",
			zap.Any("image", image),
			zap.String("dbPath", dbPath),
		)

		rowsAff := conn.FirstOrCreate(&image, "file_path = ?", file).RowsAffected

		if rowsAff > 0 {
			logger.Debug(
				"Rows affected",
				zap.Int64("rowsAff", rowsAff),
			)
		} else {
			logger.Debug(
				"No rows was added",
				zap.Int64("rowsAff", rowsAff),
			)
		}

	}

	return nil
}

func GetBounds(path string) (imageData, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0765)
	if err != nil {
		return imageData{}, err
	}
	defer f.Close()

	m, _, err := image.Decode(f)
	if err != nil {
		return imageData{}, err
	}

	fInfo, err := times.Stat(path)
	if err != nil {
		return imageData{}, err
	}

	return imageData{
		xdim:         uint(m.Bounds().Dx()),
		ydim:         uint(m.Bounds().Dy()),
		dateCreated:  fInfo.ChangeTime(),
		dateModified: fInfo.ModTime(),
	}, nil
}
