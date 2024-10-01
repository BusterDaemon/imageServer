package apis

import (
	"buster_daemon/imageserver/internal/config"
	"fmt"
	"net"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Start(cnf *config.Config, db *gorm.DB, zapper *zap.Logger) {
	app := fiber.New(
		fiber.Config{
			BodyLimit: 300 * 1024 * 1024,
		},
	)

	addRoutes(app, db, cnf, zapper)

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cnf.Address, cnf.Port))
	if err != nil {
		zapper.Fatal("Can't open listener", zap.String("address", cnf.Address), zap.Uint16("port", cnf.Port))
	}

	zapper.Error("error", zap.Error(app.Listener(ln)))
}
