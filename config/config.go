package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Cache Cache `json:"cache" yaml:"cache"`
	Proxy Proxy `json:"proxy" yaml:"proxy"`
	Limit Limit `json:"limit" yaml:"limit"`
}

type Cache struct {
	Dir string `json:"dir" yaml:"dir"`
}

type Proxy struct {
	Url string `json:"url" yaml:"url"`
}
type Limit struct {
	Rate int `json:"rate" yaml:"rate"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("读取配置: %w", err)
	}
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("解析配置: %w", err)
	}
	return config, nil
}
