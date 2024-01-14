package apis

import (
	"buster_daemon/imageserver/internal/filelist"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func getRandFile(ctx *fiber.Ctx) error {
	find := ctx.Params("contains", "")
	fl, err := filelist.GetFileList(globConf.RootFolder, find, globConf.AllowGifs)
	if err != nil {
		ctx.SendStatus(http.StatusNotFound)
		return err
	}

	f, err := filelist.GetRandomFile(fl)
	if err != nil {
		ctx.SendStatus(http.StatusNotFound)
		return err
	}

	log.Println(f)

	return ctx.SendFile(f, true)
}
