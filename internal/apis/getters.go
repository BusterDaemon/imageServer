package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func getRandFile(ctx *fiber.Ctx) error {
	name := ctx.Query("name", "")
	sLandscape := ctx.Query("orientation", "0")

	landscape, err := strconv.ParseUint(sLandscape, 10, 8)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	f, err := database.SelectRandomFile(globConf.DBPath, database.RandomParams{
		Substring: name,
		Landscape: uint8(landscape),
	})
	if err != nil {
		log.Println(err)
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendFile(f, true)
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
