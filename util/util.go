package util

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"codeberg.org/iklabib/kerat/model"
	"github.com/goccy/go-yaml"
)

func LoadConfig(configPath string) (*model.Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config model.Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func IterDir(dir string, filenames []string) ([]string, error) {
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return filenames, err
	}

	for _, v := range dirs {
		path := filepath.Join(dir, v.Name())
		if !v.IsDir() {
			filenames = append(filenames, path)
		} else {
			filenames, err = IterDir(path, filenames)
			if err != nil {
				return filenames, err
			}
		}
	}

	return filenames, nil
}

func IsNotExist(dir string) bool {
	_, err := os.Stat(dir)
	return errors.Is(err, fs.ErrNotExist)
}
