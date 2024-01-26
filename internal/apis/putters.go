package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"buster_daemon/imageserver/internal/filelist"
	"errors"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func startNewScan(ctx *fiber.Ctx) error {
	fl, err := filelist.GetFileList(globConf.RootFolder, "", globConf.AllowGifs)
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		return err
	}

	go database.InsertRecords(globConf.DBPath, fl, globLogger)
	return ctx.SendStatus(http.StatusAccepted)
}

func startNonExistScan(ctx *fiber.Ctx) error {
	var (
		filePaths []string
	)

	conn, err := gorm.Open(sqlite.Open(globConf.DBPath))
	if err != nil {
		globLogger.Error(
			"Connecting to database failed",
			zap.Error(err),
		)

		return ctx.SendStatus(http.StatusInternalServerError)
	}

	err = conn.Model(&database.Images{}).
		Pluck("file_path", &filePaths).
		Error

	if err != nil {
		globLogger.Error(
			"Error has occured",
			zap.Error(err),
		)

		return ctx.SendStatus(http.StatusNotFound)
	}

	if len(filePaths) > 0 {
		go func(files []string) {
			for _, file := range files {
				_, err := os.Stat(file)
				if errors.Is(err, os.ErrNotExist) {
					globLogger.Debug(
						"File is not exist, trying to delete it from database",
						zap.String("file", file),
						zap.String("database", globConf.DBPath),
					)

					err = database.DeleteRecord(
						globConf.DBPath,
						database.DeleteOptions{
							Id:       0,
							FilePath: file,
						},
						globLogger,
					)

					if err != nil {
						globLogger.Error(
							"Can't delete record from database",
							zap.Error(err),
						)
					}
				}
			}
		}(filePaths)
	}
	return ctx.SendStatus(http.StatusAccepted)
}
