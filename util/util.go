package util

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"codeberg.org/iklabib/kerat/model"
	"github.com/goccy/go-yaml"
)

func LoadGlobalConfig(configPath string) (*model.GlobalConfig, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config model.GlobalConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func loadSubmissionConfig(configPath string) (model.SubmissionConfig, error) {
	content, err := os.ReadFile(configPath)
	var config model.SubmissionConfig
	if err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		return config, err
	}

	return config, nil
}

func LoadSubmissionConfigs(configsDir string, enables []string) ([]model.SubmissionConfig, error) {
	var configs = []model.SubmissionConfig{}
	for _, v := range enables {
		configPath := filepath.Join(configsDir, v+".yaml")
		config, err := loadSubmissionConfig(configPath)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, nil
}

var randomizer = rand.New(rand.NewSource(10))

func RandomString() string {
	return strconv.Itoa(randomizer.Intn(6))
}
