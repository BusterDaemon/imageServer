package apis

import (
	"buster_daemon/imageserver/internal/apis/database"
	"buster_daemon/imageserver/internal/apis/tokens"
	"buster_daemon/imageserver/internal/apis/users"
	"buster_daemon/imageserver/internal/config"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func postImage(ctx *fiber.Ctx) error {
	var (
		logger     *zap.Logger    = ctx.Locals("logger").(*zap.Logger)
		config     *config.Config = ctx.Locals("conf").(*config.Config)
		db         *gorm.DB       = ctx.Locals("db").(*gorm.DB)
		username   string         = ctx.Cookies("username")
		imageData  database.Images
		imStoreDir string
		imHash     string
		fileExt    string
		filePath   string
		tagList    []string
		user       *database.User
	)

	if username == "" {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	imStoreDir = filepath.Join(
		config.RootFolder, username,
	)

	fbuf := new(bytes.Buffer)
	f, err := ctx.FormFile("file")
	if err != nil {
		logger.Error(err.Error())
		return ctx.SendStatus(http.StatusBadRequest)
	}

	tagList, err = parseTagList(
		ctx.FormValue("tags"),
	)
	if err != nil {
		logger.Error(err.Error())
		return ctx.SendStatus(http.StatusBadRequest)
	}

	contHeaders := f.Header["Content-Type"]
	switch contHeaders[0] {
	case "image/png":
		fileExt = ".png"
	case "image/jpeg":
		fileExt = ".jpeg"
	default:
		logger.Error("Unknown file format")
		return ctx.SendStatus(http.StatusBadRequest)
	}

	sf, _ := f.Open()
	_, err = io.Copy(fbuf, sf)
	if err != nil {
		logger.Error(err.Error())
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	defer sf.Close()

	sha := md5.New()
	sha.Write(fbuf.Bytes())
	imHash = fmt.Sprintf("%x", sha.Sum(nil))

	if !dirExists(imStoreDir) {
		err = os.MkdirAll(imStoreDir, 0755)
		if err != nil {
			logger.Error(err.Error())
			return ctx.SendStatus(http.StatusInternalServerError)
		}
	}

	tmp, err := os.CreateTemp(imStoreDir, "*")
	if err != nil {
		logger.Error(err.Error())
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	defer tmp.Close()

	_, err = io.Copy(tmp, fbuf)
	if err != nil {
		logger.Error(err.Error())
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	tmp.Seek(0, io.SeekStart)

	filePath = filepath.Join(
		imStoreDir,
		imHash,
	) + fileExt

	data, _, err := image.DecodeConfig(tmp)
	if err != nil {
		logger.Error(err.Error())
		os.Remove(filePath)
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	err = os.Rename(tmp.Name(), filePath)
	if err != nil {
		logger.Error(err.Error())
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	imageData = database.Images{
		FilePath:     filePath,
		XDim:         uint(data.Width),
		YDim:         uint(data.Height),
		Score:        0,
		Format:       fileExt,
		Hash:         imHash,
		DateAdded:    time.Now(),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}

	for _, tag := range tagList {
		t, err := database.GetImageTag(db, tag)
		if err != nil {
			continue
		}
		imageData.Tags = append(imageData.Tags, *t)
	}

	if len(imageData.Tags) < 1 {
		logger.Sugar().Errorf("Image from client %s has no tags", ctx.IP())
		os.Remove(filePath)
		return ctx.SendStatus(http.StatusBadRequest)
	}

	user, err = database.GetUser(db, username)
	if err != nil {
		log.Error(err)
		os.Remove(filePath)
		return ctx.SendStatus(http.StatusUnauthorized)
	}

	imageData.User = int(user.ID)

	if res := db.Create(&imageData); res.Error != nil {
		log.Error(err)
		os.Remove(filePath)
		return ctx.SendStatus(http.StatusAlreadyReported)
	}

	return ctx.SendStatus(http.StatusAccepted)
}

func registerUser(ctx *fiber.Ctx) error {
	var (
		login    string
		passw    string
		name     string
		userType int
		logger   *zap.Logger = ctx.Locals("logger").(*zap.Logger)
		db       *gorm.DB    = ctx.Locals("db").(*gorm.DB)
	)

	login = ctx.FormValue("login", "")
	passw = ctx.FormValue("password", "")
	name = ctx.FormValue("username", "")
	userType, err := strconv.Atoi(ctx.FormValue("user_type", "3"))
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	user, err := users.New(login, name, passw, userType)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	res := db.Create(&user)
	if res.Error != nil {
		logger.Error(res.Error.Error())
		return res.Error
	}
	return nil
}

func loginUser(ctx *fiber.Ctx) error {
	var (
		login    string   = ctx.FormValue("login", "")
		password string   = ctx.FormValue("password", "")
		db       *gorm.DB = ctx.Locals("db").(*gorm.DB)
	)

	if db == nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	if login == "" || password == "" {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	hashPassw, err := users.HashPassword(password)
	if err != nil {
		return ctx.SendStatus(http.StatusForbidden)
	}

	res := db.Where("login = ?", login).
		Where("passw", hashPassw).First(&database.User{})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return ctx.SendStatus(http.StatusNotFound)
	}

	err = tokens.CreateToken(login, ctx)
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	ctx.Cookie(&fiber.Cookie{
		Name:    "username",
		Value:   login,
		Expires: time.Now().Add(time.Hour * 9999),
	})

	return ctx.SendStatus(http.StatusOK)
}

func dirExists(dPath string) bool {
	_, err := os.Stat(dPath)

	return errors.Is(err, &fs.PathError{})
}

func parseTagList(tags string) ([]string, error) {
	var (
		tagList []string
		newTags []string
	)
	tagList = strings.Split(tags, ";")
	if len(tagList) < 1 {
		return nil, database.EmptyTagList{}
	}

	for _, tag := range tagList {
		tag = strings.ToLower(tag)
		tag = strings.ReplaceAll(tag, " ", "_")
		newTags = append(newTags, tag)
	}

	return newTags, nil
}
