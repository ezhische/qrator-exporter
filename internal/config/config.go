package config

import (
	"time"

	env "github.com/caarlos0/env/v9"
	"github.com/ezhische/qrator-exporter/internal/collector"
	"github.com/sirupsen/logrus"
)

type Config struct {
	APIToken string        `env:"QRATOR_X_QRATOR_AUTH,required"`
	APIURL   string        `env:"QRATOR_API_URL" envDefault:"https://api.qrator.net/request"`
	ClientID int           `env:"QRATOR_CLIENT_ID,required"`
	Domains  []int         `env:"QRATOR_DOMAINS_IDS" envSeparator:","`
	ProxyURL string        `env:"QRATOR_PROXY_URL"`
	Timeout  time.Duration `env:"QRATOR_TIMEOUT" envDefault:"5s"`
	Port     int           `env:"QRATOR_EXPORTER_PORT" envDefault:"9502"`
}

func ConfigFromEnv() (*Config, error) {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, err
	}
	return config, nil
}

func CollectorFromConfig(config *Config, logger *logrus.Logger) (*collector.Collector, error) {
	return collector.CollectorFromConfig(config.APIToken, config.ClientID, config.APIURL, config.Domains, config.ProxyURL, config.Timeout, logger)
}
