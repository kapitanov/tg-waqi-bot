package main

import (
	"github.com/kapitanov/tg-waqi-bot/pkg/api"
	"github.com/kapitanov/tg-waqi-bot/pkg/bot"
	pkgLog "github.com/kapitanov/tg-waqi-bot/pkg/log"
	"github.com/kapitanov/tg-waqi-bot/pkg/waqi"
	"github.com/spf13/viper"
	"os"
	"os/signal"
)

var log = pkgLog.New("main")

func main() {
	// Configure viper
	err := configure()
	if err != nil {
		panic(err)
	}

	// Create WAQI service adapter
	waqiService, err := waqi.NewService(
		waqi.URLOption(viper.GetString("WAQI_URL")),
		waqi.TokenOption(viper.GetString("WAQI_TOKEN")),
		waqi.CachePathOption(viper.GetString("WAQI_CACHE_PATH")),
		waqi.CacheDurationOption(viper.GetDuration("WAQI_CACHE_DURATION")))
	if err != nil {
		panic(err)
	}
	defer func() {
		err := waqiService.Close()
		if err != nil {
			panic(err)
		}
	}()

	// Create WebAPI service
	webServer, err := api.NewServer(waqiService, viper.GetString("LISTEN_ADDR"))
	if err != nil {
		panic(err)
	}

	// Create bot
	tgBot, err := bot.NewBot(
		bot.WAQIServiceOption(waqiService),
		bot.DBPathOption(viper.GetString("BOT_DB_PATH")),
		bot.URLOption(viper.GetString("TELEGRAM_API_URL")),
		bot.TokenOption(viper.GetString("TELEGRAM_API_TOKEN")),
		bot.AllowedUsernamesOption(viper.GetStringSlice("TELEGRAM_USERNAMES")))
	if err != nil {
		panic(err)
	}

	// Start everything up
	waqiService.StartUpdates()
	webServer.Start()
	err = tgBot.Start()
	if err != nil {
		panic(err)
	}

	// Wait for SIGINT
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	// Shut down everything
	tgBot.Close()
	waqiService.StopUpdates()
	webServer.Stop()

	log.Printf("goodbye")
}
