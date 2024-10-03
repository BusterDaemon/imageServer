package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"errors"
	"strings"

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
	var (
		db         *gorm.DB    = ctx.Locals("db").(*gorm.DB)
		logs       *zap.Logger = ctx.Locals("logger").(*zap.Logger)
		tagForm    string      = ctx.Query("tags", "")
		tagsParsed []string
	)

	if tagForm == "" {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	tagsParsed = strings.Split(tagForm, ";")
	if len(tagsParsed) < 1 {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	results, err := database.GetImagesWithTags(db, tagsParsed)
	if err != nil {
		logs.Error(err.Error())
		if errors.Is(err, database.NoImageFound{}) {
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		ctx.SendStatus(fiber.StatusBadRequest)
	}

	return ctx.JSON(results)
}

func getImage(ctx *fiber.Ctx) error {
	var (
		db        *gorm.DB    = ctx.Locals("db").(*gorm.DB)
		logs      *zap.Logger = ctx.Locals("logger").(*zap.Logger)
		hashImage string      = ctx.Params("hash", "")
		fileData  []string
	)
	if hashImage == "" {
		ctx.SendStatus(fiber.StatusBadRequest)
	}

	fileData, err := database.GetImagePathByHash(
		db, hashImage,
	)
	if err != nil {
		logs.Error(err.Error())
		if errors.Is(err, database.NoImageFound{}) {
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		ctx.SendStatus(fiber.StatusBadRequest)
	}

	if len(fileData) == 0 {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	ctx.Set(
		"Content-Disposition",
		"inline;filename=\""+
			fileData[1]+fileData[2]+
			"\"",
	)

	return ctx.SendFile(fileData[0])
}

func getImageInfo(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}
