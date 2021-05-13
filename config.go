package main

import (
	"os"
	"path"

	"github.com/spf13/viper"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/kapitanov/tg-waqi-bot/pkg/waqi"
)

func configure() error {
	cwd, _ := os.Getwd()
	viper.SetDefault("ENV_FILE", path.Join(cwd, ".env"))
	viper.SetDefault("WAQI_URL", waqi.DefaultURL)
	viper.SetDefault("WAQI_CACHE_DURATION", waqi.DefaultCacheDuration)
	viper.SetDefault("LISTEN_ADDR", "0.0.0.0:8000")
	viper.SetDefault("TELEGRAM_API_URL", telebot.DefaultApiURL)

	viper.AutomaticEnv()

	// $(pwd)/.env file
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}

	return nil
}
