package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"buster_daemon/imageserver/internal/filelist"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func startScan(ctx *fiber.Ctx) error {
	fl, err := filelist.GetFileList(globConf.RootFolder, "", globConf.AllowGifs)
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		return err
	}

	go database.InsertRecords(globConf.DBPath, fl)
	return ctx.SendStatus(http.StatusAccepted)
}
