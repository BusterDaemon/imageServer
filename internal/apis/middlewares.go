package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RequestLogger(db *gorm.DB, logs *zap.Logger) func() gin.HandlerFunc {
	return func() gin.HandlerFunc {
		return func(ctx *gin.Context) {
			var (
				err error
				req = database.ClientReqs{
					Ip:         ctx.ClientIP(),
					Time:       time.Now(),
					Url:        ctx.Request.RequestURI,
					Ua:         ctx.Request.UserAgent(),
					Method:     ctx.Request.Method,
					StatusCode: ctx.Copy().Writer.Status(),
				}
				header string
			)

			header = ctx.GetHeader("Dnt")
			if header == "1" {
				return
			}
			req.Queries, err = url.QueryUnescape(ctx.Request.URL.RawQuery)
			if err != nil {
				logs.Debug(req.Queries)
			}
			res := db.Model(&database.ClientReqs{}).
				Create(req)
			if res.Error != nil {
				logs.Debug(res.Error.Error())
			}
		}
	}
}
