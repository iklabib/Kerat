package util

import (
	"bytes"
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

// remove non-printable characters while keeping whitespaces
func SanitizeStdout(input []byte) []byte {
	// we want to clean control characters like this
	// \u0001\u0000\u0000\u0000\u0000\u0000\u00009
	// the last charater is sus tho, 5 digits
	return bytes.Trim(input, "\u0001\u0000\u00009\n")
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
