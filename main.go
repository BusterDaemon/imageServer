package main

import (
	"buster_daemon/imageserver/internal/filelist"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	api := app.Group("/api")
	getters := api.Group("/get")

	getters.Get("/random", func(c *fiber.Ctx) error {
		fl, err := filelist.GetFileList("/")
		if err != nil {
			log.Fatal(err)
		}

		f, err := filelist.GetRandomFile(fl)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(f)

		return c.SendFile(f)
	})

	app.Listen(":9150")
}
