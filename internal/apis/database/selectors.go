package database

import (
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

type RandomParams struct {
	Substring string
	Landscape uint8
}

type SearchParams struct {
	Substring string
	Xdim      uint
	XCompar   uint8
	XLess     uint
	Ydim      uint
	YCompar   uint8
	YLess     uint
	Landscape uint8
	Limit     uint
	Offset    uint
}

func SelectRandomFile(dbPath string, params RandomParams) (string, error) {
	var image Images

	conn, err := gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		return "", err
	}

	res := conn.Model(&image)
	if params.Substring != "" {
		res.Where("file_path LIKE ?", "%"+params.Substring+"%")
	}

	switch params.Landscape {
	case LANDSCAPE_IMAGES:
		res.Where("xdim > (ydim * 1.2)")
	case PORTRAIT_IMAGES:
		res.Where("xdim <= ydim")
	}

	res.Order("RANDOM()").First(&image)
	return image.FilePath, nil
}

func SearchImages(dbPath string, params SearchParams) ([]Images, int64, error) {
	var images []Images
	var total int64

	if params.Limit < 1 {
		params.Limit = 10
	}

	conn, err := gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		return nil, 0, err
	}

	res := conn.Model(&images)
	if params.Substring != "" {
		res.Where("file_path LIKE ?", "%"+params.Substring+"%")
	}

	if params.Landscape == LANDSCAPE_IMAGES {
		res.Where("xdim > (ydim * 1.2)")
	} else if params.Landscape == PORTRAIT_IMAGES {
		res.Where("xdim <= ydim")
	}

	if params.Xdim != 0 {
		switch params.XCompar {
		case LESS_THAN:
			res.Where("xdim < ?", params.XLess)
		case GREATER_THAN:
			res.Where("xdim > ?", params.XLess)
		case LESS_THAN_OR_EQUAL:
			res.Where("xdim <= ?", params.XLess)
		case GREATER_THAN_OR_EQUAL:
			res.Where("xdim >= ?", params.XLess)
		case EQUAL:
			res.Where("xdim = ?", params.XLess)
		}
	}

	if params.Ydim != 0 {
		switch params.YCompar {
		case LESS_THAN:
			res.Where("ydim < ?", params.YLess)
		case GREATER_THAN:
			res.Where("ydim > ?", params.YLess)
		case LESS_THAN_OR_EQUAL:
			res.Where("ydim <= ?", params.YLess)
		case GREATER_THAN_OR_EQUAL:
			res.Where("ydim >= ?", params.YLess)
		case EQUAL:
			res.Where("ydim = ?", params.YLess)
		}
	}

	err = res.Count(&total).Limit(int(params.Limit)).Offset(int(params.Offset)).Order("file_path ASC").Find(&images).Error
	if err != nil {
		return nil, 0, err
	}

	return images, total, nil
}
