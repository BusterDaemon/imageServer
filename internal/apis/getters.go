package apis

import (
	"buster_daemon/imageserver/internal/apis/database"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func getRandFile(ctx *fiber.Ctx) error {
	var (
		db    *gorm.DB    = ctx.Locals("db").(*gorm.DB)
		logs  *zap.Logger = ctx.Locals("logger").(*zap.Logger)
		image             = struct {
			filePath      string
			fileName      string
			fileExtension string
		}{}
	)
	res := db.Select("file_path", "format", "hash").
		Find(&database.Images{}).
		Order("RANDOM()").
		Limit(1)

	row := res.Row()

	err := row.Scan(
		&image.filePath,
		&image.fileExtension,
		&image.fileName,
	)
	if err != nil {
		logs.Error(err.Error())
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	ctx.Set(
		"Content-Disposition",
		"inline;filename=\""+
			image.fileName+image.fileExtension+
			"\"",
	)

	return ctx.SendFile(image.filePath)
}

func searchImages(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}

func getImage(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}

func getImageInfo(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}
