package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Address      string `mapstructure:"address"`
		ReadTimeout  string `mapstructure:"read_timeout"`
		WriteTimeout string `mapstructure:"write_timeout"`
		IdleTimeout  string `mapstructure:"idle_timeout"`
	} `mapstructure:"server"`
	Database struct {
		URL             string `mapstructure:"url"`
		MaxOpenConns    int    `mapstructure:"max_open_conns"`
		MaxIdleConns    int    `mapstructure:"max_idle_conns"`
		ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
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