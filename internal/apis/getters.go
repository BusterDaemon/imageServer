package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func getRandFile(ctx *fiber.Ctx) error {
	name := ctx.Query("name", "")
	sLandscape := ctx.Query("orientation", "0")
	sXdim := ctx.Query("xdim", "0")
	sYdim := ctx.Query("ydim", "0")
	sXCompar := ctx.Query("xcompar", "0")
	sYCompar := ctx.Query("ycompar", "0")

	xdim, err := strconv.ParseUint(sXdim, 10, 64)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	ydim, err := strconv.ParseUint(sYdim, 10, 64)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	xCompar, err := strconv.ParseUint(sXCompar, 10, 8)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	yCompar, err := strconv.ParseUint(sYCompar, 10, 8)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	landscape, err := strconv.ParseUint(sLandscape, 10, 8)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	f, err := database.SelectRandomFile(globConf.DBPath, database.RandomParams{
		Substring: name,
		Xdim:      uint(xdim),
		XCompar:   uint8(xCompar),
		Ydim:      uint(ydim),
		YCompar:   uint8(yCompar),
		Landscape: uint8(landscape),
	})
	if err != nil {
		log.Println(err)
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	fname := strings.Split(f, "/")
	ctx.Set("Content-Disposition", `inline; filename="`+fname[len(fname)-1]+`"`)

	return ctx.SendFile(f, true)
}

func specificImage(ctx *fiber.Ctx) error {
	name := ctx.Query("name", "")
	sId := ctx.Query("id", "1")
	sComp := ctx.Query("selects", "0")

	id, err := strconv.ParseUint(sId, 10, 64)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	comp, err := strconv.ParseUint(sComp, 10, 64)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	image, err := database.SelectSpecificImage(globConf.DBPath, database.SelectParams{
		Id:         uint(id),
		Name:       name,
		ComparMode: uint8(comp),
	})
	if err != nil {
		ctx.SendStatus(http.StatusNotFound)
	}

	filename := strings.Split(image.FilePath, "/")
	ctx.Set("Content-Disposition", `inline; filename="`+filename[len(filename)-1]+`"`)

	return ctx.SendFile(image.FilePath, true)
}

func searchFile(ctx *fiber.Ctx) error {
	substr := ctx.Query("name", "")
	sXdim := ctx.Query("xdim", "0")
	sYdim := ctx.Query("ydim", "0")
	sXCompar := ctx.Query("xcompar", "0")
	sYCompar := ctx.Query("ycompar", "0")
	sSortOrd := ctx.Query("sort", "0")
	sLandscape := ctx.Query("orientation", "0")
	sPage := ctx.Query("page", "1")
	sPageSize := ctx.Query("limit", "10")

	xdim, err := strconv.ParseUint(sXdim, 10, 64)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	ydim, err := strconv.ParseUint(sYdim, 10, 64)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	xCompar, err := strconv.ParseUint(sXCompar, 10, 8)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	yCompar, err := strconv.ParseUint(sYCompar, 10, 8)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	sortOrd, err := strconv.ParseUint(sSortOrd, 10, 8)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	landscape, err := strconv.ParseUint(sLandscape, 10, 8)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	page, err := strconv.ParseUint(sPage, 10, 64)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	pageSize, err := strconv.ParseUint(sPageSize, 10, 64)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	offset := (page - 1) * pageSize

	images, totalImages, err := database.SearchImages(globConf.DBPath, database.SearchParams{
		Substring: substr,
		Xdim:      uint(xdim),
		XCompar:   uint8(xCompar),
		Ydim:      uint(ydim),
		YCompar:   uint8(yCompar),
		Landscape: uint8(landscape),
		SortOrder: uint8(sortOrd),
		Limit:     uint(pageSize),
		Offset:    uint(offset),
	})
	if err != nil {
		return ctx.SendStatus(http.StatusNotFound)
	}

	if len(images) < 1 {
		return ctx.SendStatus(http.StatusNotFound)
	}

	totalPages := math.Ceil(float64(totalImages) / float64(pageSize))
	ctx.Response().Header.Add("X-TOTAL-PAGES", fmt.Sprintf("%0.f", totalPages))

	return ctx.JSON(images)
}
