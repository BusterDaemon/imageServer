package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
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
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Start(cnf *config.Config, db *gorm.DB, zapper *zap.Logger) {
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
	}),
		recover.New(),
		func(ctx *fiber.Ctx) error {
			ctx.Locals("db", db)
			ctx.Locals("conf", cnf)
			ctx.Locals("logger", zapper)
			return ctx.Next()
		},
	)

	if cnf.RateLimiter.Enable {
		api.Use(
			limiter.New(
				limiter.Config{
					Max:        int(cnf.RateLimiter.MaxRecConns),
					Expiration: time.Duration(cnf.RateLimiter.ExpirTime) * time.Minute,
					Next: func(ctx *fiber.Ctx) bool {
						if len(cnf.RateLimiter.WlIPs) > 0 {
							for _, i := range cnf.RateLimiter.WlIPs {
								if i == ctx.IP() {
									return true
								}
							}
						}
						return false
					},
				},
			),
		)
	}

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
			if ctx.Query("noCache") == "true" {
				return true
			}

			cType := ctx.GetRespHeader("Content-Type")

			for _, s := range cnf.Cache.WhitelistResp {
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

	if cnf.Logger.LogRequests {
		app.Use(func(ctx *fiber.Ctx) error {
			err := ctx.Next()

			database.InsertClientReqRecord(
				db,
				database.ClientReqs{
					Time:       time.Now(),
					Ip:         ctx.IP(),
					Url:        string(ctx.Request().URI().PathOriginal()),
					Queries:    string(ctx.Context().URI().QueryString()),
					Ua:         string(ctx.Context().UserAgent()),
					Method:     string(ctx.Request().Header.Method()),
					StatusCode: ctx.Response().StatusCode(),
				},
				zapper,
			)
			return err
		})
	}

	getters.Use(etag.New(etag.ConfigDefault))
	getters.Use(idempotency.New(idempotency.ConfigDefault))

	getters.Get("/search", searchFile)
	getters.Get("/image", specificImage)
	getters.Get("/image/info", imageInfo)
	getters.Get("/random", getRandFile)
	scanners.Put("/new", startNewScan)
	scanners.Put("/old", startNonExistScan)

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cnf.Address, cnf.Port))
	if err != nil {
		zapper.Fatal("Can't open listener", zap.String("address", cnf.Address), zap.Uint16("port", cnf.Port))
	}

	zapper.Error("error", zap.Error(app.Listener(ln)))
}
