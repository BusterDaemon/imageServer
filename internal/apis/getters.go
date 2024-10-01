package apis

import (
	"github.com/gofiber/fiber/v2"
)

func getRandFile(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
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
