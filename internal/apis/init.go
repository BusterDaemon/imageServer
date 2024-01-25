package apis

import (
	"buster_daemon/imageserver/internal/config"
	"fmt"
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
	"go.uber.org/zap"
)

var globConf *config.Config
var globLogger *zap.Logger

func Start(cnf *config.Config, zapper *zap.Logger) {
	globConf = cnf
	globLogger = zapper

	app := fiber.New()
	api := app.Group("/api")
	getters := api.Group("/get")
	putters := api.Group("/put")
	scanners := putters.Group("/scan")

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
		zapper.Debug("Enabling Basic Auth")
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
			wlContent := []string{
				"application/json",
				"text/plain",
			}

			cType := ctx.GetRespHeader("Content-Type")
			zapper.Debug("Response content type is", zap.String(cType, "Content-Type"))

			if ctx.Query("noCache") == "true" {
				return true
			}

			for _, s := range wlContent {
				if strings.Contains(cType, s) {
					zapper.Debug("Content type is in whitelist, skipping caching...")
					return true
				}
			}

			return false
		}
		zapper.Debug("Enabling caching", zap.Any("parameters", c))
		getters.Use(cache.New(c))
	}

	if cnf.Compression.UseCompression {
		zapper.Debug("Enabling compression")
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
	getters.Get("/image", specificImage)
	getters.Get("/image/info", imageInfo)
	getters.Get("/random", getRandFile)
	scanners.Put("/new", startNewScan)
	scanners.Put("/old", startNonExistScan)

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", globConf.Address, globConf.Port))
	if err != nil {
		zapper.Fatal("Can't open listener", zap.String("address", globConf.Address), zap.Uint16("port", globConf.Port))
	}

	zapper.Error("error", zap.Error(app.Listener(ln)))
}
