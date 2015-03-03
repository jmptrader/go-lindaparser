package lindaparser

import (
	"log"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

const configFilename = "config.yaml"

type Config struct {
	ECTSOverrides map[int]float32 `mapstructure:"ects_overrides"`
	Password      string          `mapstructure:"password"`
	Username      string          `mapstructure:"username"`
}

func FormatFloat(value float32, decimals int) string {
	var result string

	if decimals < 0 {
		decimals = -1
	}

	result = strconv.FormatFloat(float64(value), 'f', decimals, 32)
	result = strings.Replace(result, ".", ",", -1)

	return result
}

func JustifyLeft(input string, width int) string {
	characters := width - len([]rune(input))

	if characters < 1 {
		return input
	} else {
		return input + strings.Repeat(" ", characters)
	}
}

func LoadConfig(f func(name string) ([]byte, error)) Config {
	data, err := f(configFilename)

	if err != nil {
		log.Fatalf("Cannot read config file: %s", err)
	}

	rawConfig := make(map[string]interface{})
	yaml.Unmarshal(data, &rawConfig)

	var config Config
	err = mapstructure.Decode(rawConfig, &config)

	if err != nil {
		log.Fatalf("Cannot decode config file: %s", err)
	}

	return config
}
