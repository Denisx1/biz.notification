package nats

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	URL          string `envconfig:"URL"     required:"true" default:"localhost:4222"`
	Subject      string `envconfig:"SUBJECT"     required:"true"`
	Queue        string `envconfig:"QUEUE"     required:"true"`
	Name         string `envconfig:"NAME"     required:"true"`
	ErrorHandler func(subject string, err error)
}

func NewNatsConfig() (*Config, error) {
	var readerConfig Config
	if err := envconfig.Process("NATS", &readerConfig); err != nil {
		return &Config{}, fmt.Errorf("process envconfig:%w", err)
	}

	return &Config{
		URL:     readerConfig.URL,
		Subject: readerConfig.Subject,
		Queue:   readerConfig.Queue,
		Name:    readerConfig.Name,
	}, nil
}
func NewNatsConfigMust() *Config {
	config, err := NewNatsConfig()
	if err != nil {
		err = fmt.Errorf("failed to load nats config: %w", err)
		panic(err)
	}
	return config
}
