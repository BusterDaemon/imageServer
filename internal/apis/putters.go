package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"buster_daemon/imageserver/internal/config"
	"buster_daemon/imageserver/internal/filelist"
	"errors"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func startNewScan(ctx *fiber.Ctx) error {
	db := ctx.Locals("db").(*gorm.DB)
	logger := ctx.Locals("logger").(*zap.Logger)
	conf := ctx.Locals("conf").(*config.Config)

	fl, err := filelist.GetFileList(conf.RootFolder, "", conf.AllowGifs)
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		return err
	}

	go database.InsertRecords(
		db,
		fl,
		logger,
	)
	return ctx.SendStatus(http.StatusAccepted)
}

func startNonExistScan(ctx *fiber.Ctx) error {
	var (
		filePaths []string
	)
	db := ctx.Locals("db").(*gorm.DB)
	logger := ctx.Locals("logger").(*zap.Logger)

	err := db.Model(&database.Images{}).
		Pluck("file_path", &filePaths).
		Error

	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)

		return ctx.SendStatus(http.StatusNotFound)
	}

	if len(filePaths) > 0 {
		go func(files []string) error {
			for _, f := range files {
				_, err := os.Stat(f)
				if err != nil && errors.Is(err, os.ErrNotExist) {
					database.DeleteRecord(
						db,
						database.DeleteOptions{
							Id:       0,
							FilePath: f,
						},
						logger,
					)
				}
			}
			return nil
		}(filePaths)
	}
	return ctx.SendStatus(http.StatusAccepted)
}
