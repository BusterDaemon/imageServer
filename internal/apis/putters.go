package apis

import (
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

	// TODO: Implement writing data into database
	var Body struct {
		Files []string
	}
	Body.Files = fl
	return ctx.JSON(Body)
}
