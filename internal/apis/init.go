package apis

import (
	"buster_daemon/imageserver/internal/config"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Start(cnf *config.Config, db *gorm.DB, zapper *zap.Logger) {
	app := gin.New()
	app.MaxMultipartMemory = 300 << 20

	app.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	if cnf.Logger.LogRequests {
		zapper.Debug("Enable log requests")
		app.Use(RequestLogger(db, zapper)())
	}
	addRoutes(app, db, cnf, zapper)

	zapper.Error("error",
		zap.Error(
			http.ListenAndServe(
				fmt.Sprintf("%s:%d", cnf.Address, cnf.Port), app),
		),
	)
}
