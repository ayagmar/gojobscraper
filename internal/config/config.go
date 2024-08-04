package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Address      string        `mapstructure:"address"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	} `mapstructure:"server"`
	Database struct {
		URL         string `mapstructure:"url"`
		Name        string `mapstructure:"name"`
		MaxPoolSize int    `mapstructure:"max_pool_size"`
		MinPoolSize int    `mapstructure:"min_pool_size"`
	} `mapstructure:"database"`
	Scraper struct {
		DefaultPages int `mapstructure:"default_pages"`
	} `mapstructure:"scraper"`
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"log"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
