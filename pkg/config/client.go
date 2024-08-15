package config

import (
	"os"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type ClientConfig struct {
	Target   string `yaml:"target"`
	Key      string `yaml:"key" validate:"required,validateKeyLen"`
	User     string `yaml:"user"`
	Passwd   string `yaml:"passwd"`
	LogLevel string `yaml:"log-level"`
}

func LoadCleintConfigFile(path string) (*ClientConfig, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := new(ClientConfig)

	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return nil, err
	}

	if err := defaults.Set(cfg); err != nil {
		return nil, err
	}

	cfgValidator := validator.New()
	cfgValidator.RegisterValidation("validateKeyLen", ValidateKeyLength)

	if err := cfgValidator.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
