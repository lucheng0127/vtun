package config

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	CFG_FILE string = "config.yaml"
	USER_DB  string = "users"
)

type WebConfig struct {
	Enable bool `yaml:"enable"`
	Port   int  `yaml:"port" default:"8000"`
}

type ServerConfig struct {
	Port       int    `yaml:"port" default:"6123"`
	IP         string `yaml:"ip" validate:"required,validateIPv4Addr"`
	LogLevel   string `yaml:"log-level" default:"info"`
	Key        string `yaml:"key" validate:"required,validateKeyLen"`
	IPRange    string `yaml:"ip-range" validate:"required"`
	*WebConfig `yaml:"web"`
	Routes     []string `yaml:"routes" validate:"dive,validateCIDR"`
	PreUp      []string `yaml:"pre-up"`
	PostUp     []string `yaml:"post-up"`
	PreDown    []string `yaml:"pre-down"`
	PostDown   []string `yaml:"post-down"`
}

func GetCfgPath(dir string) string {
	return fmt.Sprintf("%s/%s", dir, CFG_FILE)
}

func GetUserDBPath(dir string) string {
	return fmt.Sprintf("%s/%s", dir, USER_DB)
}

func LoadServerConfigFile(path string) (*ServerConfig, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := new(ServerConfig)
	cfg.WebConfig = new(WebConfig)

	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return nil, err
	}

	if err := defaults.Set(cfg); err != nil {
		return nil, err
	}

	cfgValidator := validator.New()
	cfgValidator.RegisterValidation("validateKeyLen", ValidateKeyLength)
	cfgValidator.RegisterValidation("validateCIDR", ValidateCIDR)
	cfgValidator.RegisterValidation("validateIPv4Addr", ValidateIPv4Addr)

	if err := cfgValidator.Struct(cfg); err != nil {
		return nil, err
	}

	if cfg.WebConfig != nil && cfg.WebConfig.Enable {
		if err := cfgValidator.Struct(cfg.WebConfig); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
