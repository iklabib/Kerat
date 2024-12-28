package util

import (
	"archive/tar"
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

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

func TarSources(files model.SourceCode) (bytes.Buffer, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	sources := append(files.Src, files.SrcTest...)

	for _, file := range sources {
		header := &tar.Header{
			Name:    file.Filename,
			Size:    int64(len(file.SourceCode)),
			Mode:    0644,
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return buf, err
		}
		if _, err := tw.Write([]byte(file.SourceCode)); err != nil {
			return buf, err
		}
	}

	return buf, nil
}

func TarBinary(filename string, bin []byte) (bytes.Buffer, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	header := &tar.Header{
		Name:    filename,
		Size:    int64(len(bin)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return buf, err
	}
	if _, err := tw.Write(bin); err != nil {
		return buf, err
	}

	return buf, nil
}
