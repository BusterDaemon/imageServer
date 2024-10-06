package apis

import (
	"buster_daemon/imageserver/internal/config"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func addRoutes(
	app *gin.Engine,
	db *gorm.DB,
	config *config.Config,
	logs *zap.Logger,
) {
	var (
		rlStoreOptions ratelimit.InMemoryOptions
		rlOptions      ratelimit.Options
	)
	api := app.Group("/api")

	if config.RateLimiter.Enable {
		rlStoreOptions = ratelimit.InMemoryOptions{
			Rate:  time.Duration(config.RateLimiter.ExpirTime) * time.Minute,
			Limit: config.RateLimiter.MaxRecConns,
		}
		rlOptions = ratelimit.Options{
			ErrorHandler: rateLimiterErrHandler,
			KeyFunc:      rateLimitKeyFunc,
		}
		api.Use(
			ratelimit.RateLimiter(
				ratelimit.InMemoryStore(&rlStoreOptions),
				&rlOptions,
			),
		)
	}

	if config.Compression.UseCompression {
		logs.Debug("Enabling compression")
		api.Use(
			gzip.Gzip(
				int(config.Compression.CompressionLvl), gzip.WithExcludedExtensions(
					[]string{
						".webm",
						".mp4",
						".jpeg",
						".jpg",
						".png",
						".gif",
					},
				),
			),
		)
	}

	getters := api.Group("/get")
	gettersImage := getters.Group("/image")

	gettersImage.GET("/random", getRandFile(db, logs)())
	gettersImage.GET("/search", searchImages(db, logs)())
	gettersImage.GET("/:hash", getImage(db, logs)())
	gettersImage.GET("/:hash/info.json", getImageInfo(db, logs)())
}
