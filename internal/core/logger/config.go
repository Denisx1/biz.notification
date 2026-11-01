package core_logger

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// вот тут вот если в переменной окуржеения явно не задана переденная уровня логгера то тгда ставится дебаг то есть не понятно почему тот reuired
type Config struct {
	Level  string `envconfig:"LEVEL"  default:"DEBUG"`
	Folder string `envconfig:"FOLDER" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("LOGGER", &config); err != nil {
		return Config{}, fmt.Errorf("Process error config: %w", err)
	}
	return config, nil
}

func NewConfigMust() Config { // при запуску приложения нас не игтересует какая то еорректная обработка ощиблк связаных с тем что у нас нету обязательных переменных окружения которые нужны для нащего проекта то есть если их нет то мы просто вот тут вот бросим панику что бы даже не запустился проэкт
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("Get Logger config: %w", err)
		panic(err)
	}
	return config
}
