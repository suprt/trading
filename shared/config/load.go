package config

import (
	"os"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

func Load[T any](yamlPath string) (*T, error) {
	var cfg T
	if yamlPath != "" {
		data, err := os.ReadFile(yamlPath)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return &cfg, err
			}
		} else if !os.IsNotExist(err) {
			return &cfg, err
		}
	}

	if err := env.Parse(&cfg); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}
