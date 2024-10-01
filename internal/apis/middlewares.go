package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"buster_daemon/imageserver/internal/apis/tokens"
	"buster_daemon/imageserver/internal/config"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func createContext(
	db *gorm.DB,
	conf *config.Config,
	logs *zap.Logger,
) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		ctx.Locals("db", db)
		ctx.Locals("conf", conf)
		ctx.Locals("logger", logs)
		return ctx.Next()
	}
}

func createLogger() func(ctx *fiber.Ctx) error {
	return logger.New(logger.Config{
		Next:          nil,
		Done:          nil,
		Format:        "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${queryParams} | ${error} | ${ua}\n",
		TimeFormat:    "3:04:05",
		TimeZone:      "Local",
		TimeInterval:  1 * time.Second,
		Output:        os.Stdout,
		DisableColors: false,
	})
}

func createRateLimiter(
	conf *config.Config,
) func(ctx *fiber.Ctx) error {
	return limiter.New(
		limiter.Config{
			Max:        int(conf.RateLimiter.MaxRecConns),
			Expiration: time.Duration(conf.RateLimiter.ExpirTime) * time.Minute,
			Next: func(ctx *fiber.Ctx) bool {
				if len(conf.RateLimiter.WlIPs) > 0 {
					for _, i := range conf.RateLimiter.WlIPs {
						if i == ctx.IP() {
							return true
						}
					}
				}
				return false
			},
		},
	)
}

func createCaching(
	conf *config.Config,
	logs *zap.Logger,
) func(ctx *fiber.Ctx) bool {
	return func(ctx *fiber.Ctx) bool {
		if ctx.Query("noCache") == "true" {
			return true
		}

		if strings.Contains(ctx.Route().Path, "/get/image/random") {
			return true
		}

		cType := ctx.GetRespHeader("Content-Type")

		for _, s := range conf.Cache.WhitelistResp {
			if strings.Contains(cType, s) {
				logs.Debug("Content type is in whitelist, skipping caching...")
				return true
			}
		}

		return false
	}
}

func createCompressor(
	conf *config.Config,
) func(ctx *fiber.Ctx) error {
	return compress.New(
		compress.Config{
			Next:  nil,
			Level: compress.Level(conf.Compression.CompressionLvl),
		},
	)
}

func createReqsLogger(
	db *gorm.DB,
	logs *zap.Logger,
) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		headers := ctx.GetReqHeaders()
		if dnt, ok := headers["Dnt"]; ok {
			if dnt[0] == "1" {
				return ctx.Next()
			}
		}

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
			logs,
		)
		return ctx.Next()
	}
}

func createTokenVerifier() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var (
			token  string      = ctx.Cookies("token")
			rToken string      = ctx.Cookies("refreshToken")
			logger *zap.Logger = ctx.Locals("logger").(*zap.Logger)
		)

		_, err := tokens.VerifyToken(token, ctx)
		if err != nil {
			logger.Warn(err.Error())
			logger.Sugar().Warnf("Trying refresh token for user: %s", ctx.IP())
			err = tokens.RefreshToken(rToken, ctx)
			if err != nil {
				return err
			}
		}

		return ctx.Next()
	}
}
