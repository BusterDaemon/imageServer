package database

import (
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InsertRecords(dbPath string, filesPath []string) error {
	conn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	for _, file := range filesPath {
		var image Images

		err = conn.Where("file_path = ?", file).Error
		if err == nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}

		image.FilePath = file
		x, y, err := GetBounds(image.FilePath)
		if err != nil {
			return err
		}

		image.XDim = x
		image.YDim = y

		conn.FirstOrCreate(&image, "file_path = ?", file)
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
