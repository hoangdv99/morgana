package configs

import (
	"fmt"
	"os"

	"github.com/hoangdv99/morgana/cmd/configs"
	"gopkg.in/yaml.v3"
)

type ConfigFilePath string

type Config struct {
	GRPC     GRPC     `yaml:"grpc"`
	HTTP     HTTP     `yaml:"http"`
	Log      Log      `yaml:"log"`
	Auth     Auth     `yaml:"auth"`
	Database Database `yaml:"database"`
	Cache    Cache    `yaml:"cache"`
	MQ       MQ       `yaml:"mq"`
	Download Download `yaml:"download"`
}

func NewConfig(filePath ConfigFilePath) (Config, error) {
	var (
		configBytes = configs.DefaultConfigBytes
		config      = Config{}
		err         error
	)

	if filePath != "" {
		configBytes, err = os.ReadFile(string(filePath))
		if err != nil {
			return Config{}, fmt.Errorf("failed to read config file %s: %w", filePath, err)
		}
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return config, nil
}
