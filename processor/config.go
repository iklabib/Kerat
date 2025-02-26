package processor

import (
	"os"

	"codeberg.org/iklabib/kerat/processor/types"
	"github.com/goccy/go-yaml"
)

func LoadConfig(configPath string) (*types.Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config types.Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
