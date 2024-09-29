package tokens

import (
	"buster_daemon/imageserver/internal/config"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type unknownSigningMethod struct{}
type invalidToken struct{}

func (e unknownSigningMethod) Error() string {
	return "Unknown signing method"
}

func (e invalidToken) Error() string {
	return "Invalid token"
}

func CreateToken(login string, appCtx *fiber.Ctx) error {
	var (
		expireToken   int64          = time.Now().Add(time.Minute * 15).Unix()
		expireRefresh int64          = time.Now().Add(time.Hour * 24).Unix()
		timeNow       int64          = time.Now().Unix()
		logger        *zap.Logger    = appCtx.Locals("logger").(*zap.Logger)
		config        *config.Config = appCtx.Locals("conf").(*config.Config)
		tCookie       fiber.Cookie
		rtCookie      fiber.Cookie
	)
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": login,
		"iss": "imageServer",
		"aud": "users",
		"exp": expireToken,
		"iat": timeNow,
	})

	key, err := config.ReadKey()
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	t, err := token.SignedString(key)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"iss": "imageServer",
		"exp": expireRefresh,
		"sub": login,
		"aud": "users",
		"iat": timeNow,
	})
	rt, err := refresh.SignedString(key)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	tCookie.Expires = time.Unix(expireToken, 0)
	tCookie.Name = "token"
	tCookie.Path = "/api"
	tCookie.Value = t

	rtCookie.Expires = time.Unix(expireRefresh, 0)
	rtCookie.Name = "refreshToken"
	rtCookie.Path = "/api"
	rtCookie.Value = rt

	appCtx.Cookie(&tCookie)
	appCtx.Cookie(&rtCookie)

	return nil
}

func VerifyToken(token string, appCtx *fiber.Ctx) (*jwt.Token, error) {
	var (
		conf   *config.Config = appCtx.Locals("conf").(*config.Config)
		logger *zap.Logger    = appCtx.Locals("logger").(*zap.Logger)
	)

	key, err := conf.ReadKey()
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	tok, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, unknownSigningMethod{}
		}
		return key, nil
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if !tok.Valid {
		return tok, appCtx.SendStatus(http.StatusUnauthorized)
	}

	return tok, nil
}

func RefreshToken(refresh string, appCtx *fiber.Ctx) error {
	var (
		config *config.Config = appCtx.Locals("conf").(*config.Config)
		logger *zap.Logger    = appCtx.Locals("logger").(*zap.Logger)
		aToken fiber.Cookie   = fiber.Cookie{Name: "token", Path: "/api"}
		key    []byte
	)

	key, err := config.ReadKey()
	if err != nil {
		logger.Error(err.Error())
		return appCtx.SendStatus(http.StatusInternalServerError)
	}

	tok, err := jwt.Parse(refresh, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, err
		}

		return key, nil
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	_, ok := tok.Claims.(jwt.MapClaims)
	if !ok && tok.Valid {
		logger.Error("Invalid token")
		return invalidToken{}
	}

	parser := jwt.Parser{}
	claims := jwt.MapClaims{}
	_, _, err = parser.ParseUnverified(refresh, &claims)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	aToken.Value, err = token.SignedString(key)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	aToken.Expires = time.Unix(claims["iat"].(int64), 0)
	appCtx.Cookie(&aToken)

	return nil
}
