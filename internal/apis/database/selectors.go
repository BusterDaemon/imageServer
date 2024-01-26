package database

import (
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	ALL_IMAGES uint8 = iota
	PORTRAIT_IMAGES
	LANDSCAPE_IMAGES
)

const (
	LESS_THAN uint8 = iota
	GREATER_THAN
	LESS_THAN_OR_EQUAL
	GREATER_THAN_OR_EQUAL
	EQUAL
)

const (
	AND uint8 = iota
	OR
)

const (
	NAME_ASC uint8 = iota
	NAME_DESC
	ID_ASC
	ID_DESC
	XDIM_ASC
	XDIM_DESC
	YDIM_ASC
	YDIM_DESC
)

type RandomParams struct {
	Substring string
	Xdim      uint
	XCompar   uint8
	Ydim      uint
	YCompar   uint8
	Landscape uint8
}

type SelectParams struct {
	Id         uint
	Name       string
	ComparMode uint8
}

type SearchParams struct {
	Substring string
	Xdim      uint
	XCompar   uint8
	Ydim      uint
	YCompar   uint8
	Landscape uint8
	SortOrder uint8
	Limit     uint
	Offset    uint
}

func SelectRandomFile(dbPath string, params RandomParams, logger *zap.Logger) (string, error) {
	var image Images

	logger.Info(
		"Connecting to database",
		zap.String("dbPath", dbPath),
	)

	conn, err := gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		logger.Error(
			"Can't connect to database",
			zap.Error(err),
		)
		return "", err
	}

	res := conn.Model(&image)
	if params.Substring != "" {
		logger.Debug(
			"Searching file with substring",
			zap.String("substring", params.Substring),
		)
		res = res.Where("file_path LIKE ?", "%"+params.Substring+"%")
	}

	if params.Xdim != 0 {
		switch params.XCompar {
		case LESS_THAN:
			res = res.Where("xdim < ?", params.Xdim)
		case GREATER_THAN:
			res = res.Where("xdim > ?", params.Xdim)
		case LESS_THAN_OR_EQUAL:
			res = res.Where("xdim <= ?", params.Xdim)
		case GREATER_THAN_OR_EQUAL:
			res = res.Where("xdim >= ?", params.Xdim)
		case EQUAL:
			res = res.Where("xdim = ?", params.Xdim)
		}
	}

	if params.Ydim != 0 {
		switch params.XCompar {
		case LESS_THAN:
			res = res.Where("ydim < ?", params.Ydim)
		case GREATER_THAN:
			res = res.Where("ydim > ?", params.Ydim)
		case LESS_THAN_OR_EQUAL:
			res = res.Where("ydim <= ?", params.Ydim)
		case GREATER_THAN_OR_EQUAL:
			res = res.Where("ydim >= ?", params.Ydim)
		case EQUAL:
			res = res.Where("ydim = ?", params.Ydim)
		}
	}

	switch params.Landscape {
	case LANDSCAPE_IMAGES:
		logger.Debug(
			"Searching a landscape oriented image",
		)
		res = res.Where("xdim > (ydim * 1.2)")
	case PORTRAIT_IMAGES:
		logger.Debug(
			"Searching a portrait oriented image",
		)
		res = res.Where("xdim <= ydim")
	}

	err = res.Order("RANDOM()").First(&image).Error
	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)
		return "", err
	}

	logger.Debug(
		"Found the image",
		zap.Any("image", image),
	)

	return image.FilePath, nil
}

func SelectSpecificImage(dbPath string, params SelectParams, logger *zap.Logger) (Images, error) {
	var image Images

	logger.Info(
		"Connecting to database",
		zap.String("dbPath", dbPath),
	)

	conn, err := gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)
		return Images{}, err
	}
	res := conn.Model(&image)

	logger.Debug(
		"Searching image with parameters",
		zap.Any("params", params),
	)

	if params.Id != 0 && params.Name != "" {
		switch params.ComparMode {
		case AND:
			res = res.Where("file_path LIKE ?", "%"+params.Name+"%").Where("id = ?", params.Id)
		case OR:
			res = res.Where("file_path LIKE ?", "%"+params.Name+"%").Or("id = ?", params.Id)
		default:
			res = res.Where("file_path LIKE ?", "%"+params.Name+"%").Or("id = ?", params.Id)
		}
	} else if params.Id == 0 && params.Name != "" {
		res = res.Where("file_path LIKE ?", "%"+params.Name+"%")
	} else if params.Id != 0 && params.Name == "" {
		res = res.Where("id = ?", params.Id)
	}

	err = res.First(&image).Error
	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)
		return Images{}, err
	}

	logger.Debug(
		"Found the image",
		zap.Any("image", image),
	)
	return image, nil
}

func SearchImages(dbPath string, params SearchParams, logger *zap.Logger) ([]Images, int64, error) {
	var (
		images []Images
		total  int64
		order  string
	)

	if params.Limit < 1 {
		params.Limit = 10
	}

	logger.Info(
		"Connecting to database",
		zap.String("dbPath", dbPath),
	)

	conn, err := gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)
		return nil, 0, err
	}

	logger.Debug(
		"Searching image with parameters",
		zap.Any("params", params),
	)

	res := conn.Model(&images)
	if params.Substring != "" {
		res = res.Where("file_path LIKE ?", "%"+params.Substring+"%")
	}

	if params.Landscape == LANDSCAPE_IMAGES {
		res = res.Where("xdim > (ydim * 1.2)")
	} else if params.Landscape == PORTRAIT_IMAGES {
		res = res.Where("xdim <= ydim")
	}

	if params.Xdim != 0 {
		switch params.XCompar {
		case LESS_THAN:
			res = res.Where("xdim < ?", params.Xdim)
		case GREATER_THAN:
			res = res.Where("xdim > ?", params.Xdim)
		case LESS_THAN_OR_EQUAL:
			res = res.Where("xdim <= ?", params.Xdim)
		case GREATER_THAN_OR_EQUAL:
			res = res.Where("xdim >= ?", params.Xdim)
		case EQUAL:
			res = res.Where("xdim = ?", params.Xdim)
		}
	}

	if params.Ydim != 0 {
		switch params.YCompar {
		case LESS_THAN:
			res = res.Where("ydim < ?", params.Ydim)
		case GREATER_THAN:
			res = res.Where("ydim > ?", params.Ydim)
		case LESS_THAN_OR_EQUAL:
			res = res.Where("ydim <= ?", params.Ydim)
		case GREATER_THAN_OR_EQUAL:
			res = res.Where("ydim >= ?", params.Ydim)
		case EQUAL:
			res = res.Where("ydim = ?", params.Ydim)
		}
	}

	switch params.SortOrder {
	case NAME_ASC:
		order = "file_path ASC"
	case NAME_DESC:
		order = "file_path DESC"
	case ID_ASC:
		order = "id ASC"
	case ID_DESC:
		order = "id DESC"
	case XDIM_ASC:
		order = "xdim ASC"
	case XDIM_DESC:
		order = "xdim DESC"
	case YDIM_ASC:
		order = "ydim ASC"
	case YDIM_DESC:
		order = "ydim DESC"
	default:
		order = "file_path ASC"
	}

	err = res.Count(&total).
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Order(order).
		Find(&images).
		Error
	if err != nil {
		logger.Error(
			"Error has occured",
			zap.Error(err),
		)
		return nil, 0, err
	}

	return images, total, nil
}
