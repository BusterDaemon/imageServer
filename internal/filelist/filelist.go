package filelist

import (
	"errors"
	"io/fs"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

func GetFileList(root []string, find string, gif bool) ([]string, error) {
	var list []string
	for _, r := range root {
		filepath.WalkDir(r, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && (strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") || (gif && strings.HasSuffix(path, ".gif"))) {
				if strings.Contains(strings.ToLower(path), strings.ToLower(find)) {
					list = append(list, path)
				}
			}
			return nil
		})
	}

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
