package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	Application struct {
		Http struct {
			Timeout int `mapstructure:"timeout"`
		} `mapstructure:"http"`
		Log struct {
			Level string `mapstructure:"level"`
		} `mapstructure:"log"`
		Proxy struct {
			Target string `mapstructure:"target"`
		} `mapstructure:"proxy"`
	} `mapstructure:"application"`
}

var (
	appConfig Config
	once      sync.Once
)

func LoadConfig() {
	once.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("config")

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config file, %s", err)
		}

		if err := viper.Unmarshal(&appConfig); err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}
	})
}

func GetConfig() Config {
	return appConfig
}
