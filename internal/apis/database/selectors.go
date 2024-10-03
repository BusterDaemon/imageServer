package database

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type NoImageFound struct{}

type ImageWithTags struct {
	XDim       int       `json:"width"`
	YDim       int       `json:"height"`
	AuthorName string    `json:"author"`
	Score      int       `json:"score"`
	TagList    string    `json:"tags"`
	Hash       string    `json:"md5hash"`
	Created    time.Time `json:"created_at"`
	Modified   time.Time `json:"modified_at"`
}

func (e NoImageFound) Error() string {
	return "Can't find any image"
}

func GetImageTag(db *gorm.DB, tagName string) (*Tag, error) {
	var (
		tag Tag = Tag{Value: tagName}
	)
	res := db.Where("value = ?", tagName).FirstOrCreate(&tag)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, res.Error
	}

	return &tag, nil
}

func GetUser(db *gorm.DB, username string) (*User, error) {
	var (
		usvr User
	)

	if res := db.Where("login = ?", username).First(&usvr); res.Error != nil {
		return nil, res.Error
	}

	return &usvr, nil
}

func GetImagesWithTags(db *gorm.DB, tags []string) ([]ImageWithTags, error) {
	var (
		tempRes ImageWithTags
		results []ImageWithTags = []ImageWithTags{}
	)

	// TODO: Rewrite to GORM
	resdb := db.Raw(
		`SELECT DISTINCT i.xdim, i.ydim,
u.display_name, i.score,
(SELECT GROUP_CONCAT(t.value, ';')
FROM tags t INNER JOIN image_tags it2 ON t.id = it2.tag_id
WHERE it2.images_id = i.id
ORDER BY t.value) AS tags,
i.hash, i.created_at, i.modified_at
FROM images i
INNER JOIN image_tags it ON it.images_id = i.id
INNER JOIN users u ON i."user" = u.id
INNER JOIN tags t ON t.id = it.tag_id
WHERE t.value IN (?)
GROUP BY i.id;`, tags,
	)
	res, err := resdb.Rows()
	if err != nil {
		return nil, err
	}

	for res.Next() {
		err = res.Scan(
			&tempRes.XDim,
			&tempRes.YDim,
			&tempRes.AuthorName,
			&tempRes.Score,
			&tempRes.TagList,
			&tempRes.Hash,
			&tempRes.Created,
			&tempRes.Modified,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, tempRes)
	}

	if len(results) < 1 {
		return nil, NoImageFound{}
	}

	return results, nil
}

func GetImagePathByHash(
	db *gorm.DB,
	hash string,
) ([]string, error) {
	var (
		path   string
		name   string
		format string
	)

	res := db.Select("file_path, hash, format").
		Where("hash = ?", hash).
		Find(&Images{})
	if res.Error != nil {
		return nil, res.Error
	}

	err := res.Row().Scan(&path, &name, &format)
	if err != nil {
		return nil, err
	}

	return []string{
		path,
		name,
		format,
	}, nil
}

func GetImageInfoByHash(
	db *gorm.DB,
	hash string,
) (ImageWithTags, error) {
	var imData ImageWithTags
	res := db.
		Table("images").
		Select(
			"images.xdim",
			"images.ydim",
			"images.score",
			"GROUP_CONCAT(DISTINCT tags.value) as tags",
			"users.display_name",
			"images.hash",
			"images.created_at",
			"images.modified_at",
		).
		Joins("INNER JOIN users ON users.id = images.user").
		Joins("INNER JOIN image_tags ON images.id = image_tags.images_id").
		Joins("INNER JOIN tags ON tags.id = image_tags.tag_id").
		Where("hash = ?", hash)
	if res.Error != nil {
		return ImageWithTags{}, res.Error
	}

	err := res.Row().Scan(
		&imData.XDim,
		&imData.YDim,
		&imData.Score,
		&imData.TagList,
		&imData.AuthorName,
		&imData.Hash,
		&imData.Created,
		&imData.Modified,
	)
	if err != nil {
		return ImageWithTags{}, err
	}

	imData.TagList = strings.ReplaceAll(
		imData.TagList, ",", ";",
	)

	return imData, nil
}
