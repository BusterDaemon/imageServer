package apis

import (
	"buster_daemon/imageserver/internal/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func addRoutes(
	app *fiber.App,
	db *gorm.DB,
	config *config.Config,
	logs *zap.Logger,
) {
	api := app.Group("/api")
	getters := api.Group("/get")
	posters := api.Group("/post")

	api.Use(
		createLogger(),
		recover.New(),
		createContext(db, config, logs),
	)

	if config.RateLimiter.Enable {
		api.Use(createRateLimiter(config))
	}

	if config.Cache.UseCache {
		c := cache.ConfigDefault
		c.Expiration = time.Duration(config.Cache.ExpCache) * time.Second
		c.Next = createCaching(config, logs)
		logs.Debug("Enabling caching", zap.Any("parameters", c))
		getters.Use(cache.New(c))
	}

	if config.Compression.UseCompression {
		logs.Debug("Enabling compression")
		getters.Use(createCompressor(config))
	}

	if config.Logger.LogRequests {
		app.Use(createReqsLogger(db, logs))
	}

	getters.Use(etag.New(etag.ConfigDefault))
	getters.Use(idempotency.New(idempotency.ConfigDefault))

	getterImager := getters.Group("/image")

	getterImager.Get("/random", getRandFile)
	getterImager.Get("/search", searchImages)
	getterImager.Get("/:hash", getImage)
	getterImager.Get("/:hash/info.json", getImageInfo)
	posters.Post("/login", loginUser)

	posters.Use(createTokenVerifier())
	posters.Post("/new", postImage)
	posters.Post("/register", registerUser)
}
