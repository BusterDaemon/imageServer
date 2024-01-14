package main

import (
	"buster_daemon/imageserver/internal/apis"
	"buster_daemon/imageserver/internal/config"

	"github.com/gofiber/fiber/v2/log"
)

func main() {
	conf, err := config.ReadConfig("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	apis.Start(&conf)
}
