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

func SelectSpecificImage(dbPath string, params SelectParams) (Images, error) {
	var image Images
	conn, err := gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		return Images{}, err
	}
	res := conn.Model(&image)

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
		return Images{}, err
	}

	return image, nil
}

func SearchImages(dbPath string, params SearchParams) ([]Images, int64, error) {
	var (
		images []Images
		total  int64
		order  string
	)

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
			res.Where("xdim < ?", params.Xdim)
		case GREATER_THAN:
			res.Where("xdim > ?", params.Xdim)
		case LESS_THAN_OR_EQUAL:
			res.Where("xdim <= ?", params.Xdim)
		case GREATER_THAN_OR_EQUAL:
			res.Where("xdim >= ?", params.Xdim)
		case EQUAL:
			res.Where("xdim = ?", params.Xdim)
		}
	}

	if params.Ydim != 0 {
		switch params.YCompar {
		case LESS_THAN:
			res.Where("ydim < ?", params.Ydim)
		case GREATER_THAN:
			res.Where("ydim > ?", params.Ydim)
		case LESS_THAN_OR_EQUAL:
			res.Where("ydim <= ?", params.Ydim)
		case GREATER_THAN_OR_EQUAL:
			res.Where("ydim >= ?", params.Ydim)
		case EQUAL:
			res.Where("ydim = ?", params.Ydim)
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

	err = res.Count(&total).Limit(int(params.Limit)).Offset(int(params.Offset)).Order(order).Find(&images).Error
	if err != nil {
		return nil, 0, err
	}

	return images, total, nil
}
