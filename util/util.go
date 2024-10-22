package util

import (
	"math/rand"
	"os"
	"strconv"
	"unicode"

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

var randomizer = rand.New(rand.NewSource(10))

func RandomString() string {
	return strconv.Itoa(randomizer.Intn(6))
}

// remove non-printable characters while keeping whitespaces
func SanitizeStdout(input []byte) []byte {
	var result []byte
	for _, ch := range input {
		r := rune(ch)
		if !unicode.IsSpace(r) && !unicode.IsPrint(r) {
			continue
		}

		result = append(result, ch)
	}
	return result
}

// clean from start until printable character found
// clean from end until printable character found
func CleanJson(input []byte) []byte {
	start := 0
	for start < len(input) && unicode.IsControl(rune(input[start])) {
		start++
	}

	end := len(input) - 1
	for end > start && unicode.IsControl(rune(input[end])) {
		end--
	}

	return input[start : end+1]
}
