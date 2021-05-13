package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/spf13/viper"

	"github.com/kapitanov/tg-waqi-bot/pkg/api"
	"github.com/kapitanov/tg-waqi-bot/pkg/bot"
	"github.com/kapitanov/tg-waqi-bot/pkg/waqi"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetOutput(os.Stderr)

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
		waqi.CacheDurationOption(viper.GetDuration("WAQI_CACHE_DURATION")),
		waqi.LoggerOption(log.New(log.Writer(), "waqi: ", log.Flags())))
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
	webServer, err := api.NewServer(
		waqiService,
		viper.GetString("LISTEN_ADDR"),
		log.New(log.Writer(), "api: ", log.Flags()))
	if err != nil {
		panic(err)
	}

	// Create bot
	tgBot, err := bot.NewBot(
		bot.WAQIServiceOption(waqiService),
		bot.DBPathOption(viper.GetString("BOT_DB_PATH")),
		bot.URLOption(viper.GetString("TELEGRAM_API_URL")),
		bot.TokenOption(viper.GetString("TELEGRAM_API_TOKEN")),
		bot.AllowedUsernamesOption(viper.GetStringSlice("TELEGRAM_USERNAMES")),
		bot.LoggerOption(log.New(log.Writer(), "bot: ", log.Flags())))
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
