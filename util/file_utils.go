package util

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func IsNotExist(dir string) bool {
	_, err := os.Stat(dir)
	return errors.Is(err, fs.ErrNotExist)
}

func GetFilenames(dir string, filenames []string) ([]string, error) {
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return filenames, err
	}

	for _, v := range dirs {
		path := filepath.Join(dir, v.Name())
		if !v.IsDir() {
			filenames = append(filenames, path)
		} else {
			filenames, err = GetFilenames(path, filenames)
			if err != nil {
				return filenames, err
			}
		}
	}

	return filenames, nil
}
