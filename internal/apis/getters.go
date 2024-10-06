package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func getRandFile(db *gorm.DB, logs *zap.Logger) func() gin.HandlerFunc {
	return func() gin.HandlerFunc {
		return func(ctx *gin.Context) {
			var (
				image = struct {
					filePath      string
					fileName      string
					fileExtension string
				}{}
			)
			res := db.Select("file_path", "format", "hash").
				Model(&database.Images{}).
				Order("RANDOM()").
				Limit(1)

			row := res.Row()

			err := row.Scan(
				&image.filePath,
				&image.fileExtension,
				&image.fileName,
			)
			if err != nil {
				logs.Error(err.Error())
				ctx.Status(
					http.StatusInternalServerError,
				)
				return
			}

			ctx.Header(
				"Content-Disposition",
				"inline;filename=\""+
					image.fileName+image.fileExtension+
					"\"",
			)
			logs.Debug("Image file: ", zap.Any("image", image.filePath))

			ctx.File(image.filePath)
		}
	}
}

func searchImages(db *gorm.DB, logs *zap.Logger) func() gin.HandlerFunc {
	return func() gin.HandlerFunc {
		return func(ctx *gin.Context) {
			var (
				tagForm    string = ctx.Query("tags")
				tagsParsed []string
			)

			if tagForm == "" {
				ctx.Status(http.StatusBadRequest)
				return
			}

			tagsParsed = strings.Split(tagForm, ";")
			if len(tagsParsed) < 1 {
				ctx.Status(http.StatusBadRequest)
				return
			}

			results, err := database.GetImagesWithTags(db, tagsParsed)
			if err != nil {
				logs.Error(err.Error())
				if errors.Is(err, database.NoImageFound{}) {
					ctx.Status(http.StatusNotFound)
					return
				}
				ctx.Status(fiber.StatusBadRequest)
			}

			ctx.JSON(http.StatusOK, results)
		}
	}
}

func getImage(db *gorm.DB, logs *zap.Logger) func() gin.HandlerFunc {
	return func() gin.HandlerFunc {
		return func(ctx *gin.Context) {
			var (
				hashImage string = ctx.Param("hash")
				fileData  []string
			)
			if hashImage == "" {
				ctx.Status(http.StatusBadRequest)
				return
			}

			fileData, err := database.GetImagePathByHash(
				db, hashImage,
			)
			if err != nil {
				logs.Error(err.Error())
				if errors.Is(err, database.NoImageFound{}) {
					ctx.Status(http.StatusNotFound)
					return
				}
				ctx.Status(http.StatusBadRequest)
				return
			}

			if len(fileData) == 0 {
				ctx.Status(http.StatusNotFound)
				return
			}

			ctx.Header(
				"Content-Disposition",
				"inline;filename=\""+
					fileData[1]+fileData[2]+
					"\"",
			)

			ctx.File(fileData[0])
		}
	}
}

func getImageInfo(db *gorm.DB, logs *zap.Logger) func() gin.HandlerFunc {
	return func() gin.HandlerFunc {
		return func(ctx *gin.Context) {
			var (
				fileHash string = ctx.Param("hash")
			)
			if fileHash == "" {
				ctx.Status(http.StatusBadRequest)
				return
			}

			info, err := database.GetImageInfoByHash(db, fileHash)
			if err != nil {
				logs.Error(err.Error())
				ctx.Status(http.StatusBadRequest)
				return
			}

			ctx.JSON(http.StatusOK, info)
		}
	}
}
