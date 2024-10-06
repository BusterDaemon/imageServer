package apis

import (
	"buster_daemon/imageserver/internal/config"

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

	api := app.Group("/api")
	getters := api.Group("/get")
	gettersImage := getters.Group("/image")

	gettersImage.GET("/random", getRandFile(db, logs)())
	gettersImage.GET("/search", searchImages(db, logs)())
	gettersImage.GET("/:hash", getImage(db, logs)())
	gettersImage.GET("/:hash/info.json", getImageInfo(db, logs)())

	// if config.RateLimiter.Enable {
	// 	api.Use(createRateLimiter(config))
	// }

	// if config.Cache.UseCache {
	// 	c := cache.ConfigDefault
	// 	c.Expiration = time.Duration(config.Cache.ExpCache) * time.Second
	// 	c.Next = createCaching(config, logs)
	// 	logs.Debug("Enabling caching", zap.Any("parameters", c))
	// 	getters.Use(cache.New(c))
	// }

	// if config.Compression.UseCompression {
	// 	logs.Debug("Enabling compression")
	// 	getters.Use(createCompressor(config))
	// }

	// if config.Logger.LogRequests {
	// 	app.Use(createReqsLogger(db, logs))
	// }

	// app.Route("/api",
	// 	func(r chi.Router) {
	// 		r.Route("/get", func(r chi.Router) {
	// 			r.Route("/image", func(r chi.Router) {
	// 				r.Get("/random", func(w http.ResponseWriter, r *http.Request) {
	// 					w.Write([]byte("Hello World"))
	// 				})
	// 				r.Get("/search", func(w http.ResponseWriter, r *http.Request) {
	// 					w.Write([]byte("Search"))
	// 				})
	// 				r.Get("/{hash}", func(w http.ResponseWriter, r *http.Request) {
	// 					v := chi.URLParam(r, "hash")
	// 					w.Write([]byte(v))
	// 				})
	// 				r.Get("/{hash}/info.json", func(w http.ResponseWriter, r *http.Request) {
	// 					v := chi.URLParam(r, "hash")
	// 					w.Write([]byte(v + ".json"))
	// 				})
	// 			})
	// 		})
	// 	})

	// getters.Use(etag.New(etag.ConfigDefault))
	// getters.Use(idempotency.New(idempotency.ConfigDefault))

	// getterImager := getters.Group("/image")

	// getterImager.Get("/random", getRandFile)
	// getterImager.Get("/search", searchImages)
	// getterImager.Get("/:hash", getImage)
	// getterImager.Get("/:hash/info.json", getImageInfo)

	// posters.Post("/login", loginUser)
	// posters.Post("/register", registerUser)

	// posters.Use(createTokenVerifier())
	// posters.Post("/new", postImage)
}
