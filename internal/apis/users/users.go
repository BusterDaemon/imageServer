package users

import (
	"buster_daemon/imageserver/internal/apis/database"
	"crypto/sha512"
	"fmt"
)

const (
	TYPE_USER_ADMIN = iota
	TYPE_USER_MODERATOR
	TYPE_USER_SIMPLE
)

type InvalidUserData struct{}

func (e InvalidUserData) Error() string {
	return "Invalid user data"
}

func New(login string, name string, passw string, userType int) (database.User, error) {
	var (
		userPassword string
	)
	if login == "" || name == "" || passw == "" {
		return database.User{}, &InvalidUserData{}
	}

	userPassword, err := HashPassword(passw)
	if err != nil {
		return database.User{}, err
	}

	return database.User{
		Login:       login,
		DisplayName: name,
		Passw:       userPassword,
		UserType:    userType,
	}, nil
}

func HashPassword(pwd string) (string, error) {
	hasher := sha512.New()
	_, err := hasher.Write([]byte(pwd))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
