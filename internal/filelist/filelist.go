package filelist

import (
	"errors"
	"io/fs"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

func GetFileList(root string) ([]string, error) {
	var list []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && (strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg")) {
			list = append(list, path)
		}
		return nil
	})

	if len(list) < 1 {
		return nil, errors.New("no images was found")
	}

	return list, nil
}

func GetRandomFile(fileP []string) (string, error) {
	if len(fileP) < 1 {
		return "", errors.New("string array is empty")
	}

	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	return fileP[rng.Intn(len(fileP))], nil
}
