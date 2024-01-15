package apis

import (
	"buster_daemon/imageserver/internal/config"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var globConf *config.Config

func Start(cnf *config.Config) {
	globConf = cnf

	app := fiber.New()
	api := app.Group("/api")
	getters := api.Group("/get")
	putters := api.Group("/put")

	api.Use(logger.New(logger.Config{
		Next:          nil,
		Done:          nil,
		Format:        "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${queryParams} | ${error} | ${ua}\n",
		TimeFormat:    "3:04:05",
		TimeZone:      "Local",
		TimeInterval:  1 * time.Second,
		Output:        os.Stdout,
		DisableColors: false,
	}))
	api.Use(recover.New())

	if cnf.Auth.Enable {
		api.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				cnf.Auth.Login: cnf.Auth.Password,
			},
		}))
	}

	if cnf.Cache.UseCache {
		c := cache.ConfigDefault
		c.Expiration = time.Duration(cnf.Cache.ExpCache) * time.Second
		c.Next = func(ctx *fiber.Ctx) bool {
			if ctx.Query("noCache") == "true" || strings.Contains(ctx.OriginalURL(), "/search") {
				return true
			} else {
				return false
			}
		}
		getters.Use(cache.New(c))
	}
	if cnf.Compression.UseCompression {
		getters.Use(compress.New(
			compress.Config{
				Next:  nil,
				Level: compress.Level(cnf.Compression.CompressionLvl),
			},
		))
	}

	getters.Use(etag.New(etag.ConfigDefault))
	getters.Use(idempotency.New(idempotency.ConfigDefault))

	getters.Get("/search", searchFile)
	getters.Get("/random", getRandFile)
	putters.Put("/scan", startScan)

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", globConf.Address, globConf.Port))
	if err != nil {
		log.Fatal(err)
	}

	app.Listener(ln)
}
