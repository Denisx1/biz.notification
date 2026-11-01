package core_config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TimeZone *time.Location `envconfig:"TIMEZONE"     required:"true"`
	BaseSMS  string         `envconfig:"SMS_TEXT"      required:"true"`
	APIKey   string         `envconfig:"API_KEY"       required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("APP", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig:%w", err)
	}
	return config, nil
}
func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get Postgres connection pool: %w", err)
		panic(err)
	}
	return config
}
