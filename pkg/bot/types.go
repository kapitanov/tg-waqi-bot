package bot

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"gopkg.in/tucnak/telebot.v2"

	"github.com/kapitanov/tg-waqi-bot/pkg/waqi"
)

type options struct {
	URL              string
	Token            string
	DBPath           string
	WAQI             waqi.Service
	AllowedUsernames []string
	Logger           *log.Logger
}

// Option is a configuration option for NewBot function
type Option func(*options)

// WAQIServiceOption sets WAQI service instance
func WAQIServiceOption(waqiService waqi.Service) Option {
	return func(opts *options) {
		opts.WAQI = waqiService
	}
}

// URLOption sets Telegram BotAPI URL
func URLOption(url string) Option {
	return func(opts *options) {
		opts.URL = url
	}
}

// TokenOption sets Telegram access token
func TokenOption(token string) Option {
	return func(opts *options) {
		opts.Token = token
	}
}

// DBPathOption sets path to DB file
func DBPathOption(path string) Option {
	return func(opts *options) {
		opts.DBPath = path
	}
}

// AllowedUsernamesOption sets list of allowed usernames
func AllowedUsernamesOption(allowedUsernames []string) Option {
	return func(opts *options) {
		opts.AllowedUsernames = allowedUsernames
	}
}

// LoggerOption sets logger instance
func LoggerOption(logger *log.Logger) Option {
	return func(opts *options) {
		opts.Logger = logger
	}
}

// Bot wraps Bot logic
type Bot interface {
	// Start starts Bot
	Start() error

	// Close shuts down Bot
	Close()
}

// NewBot creates an instance of Bot
func NewBot(fn ...Option) (Bot, error) {
	// Generate options
	opts := &options{
		URL:    telebot.DefaultApiURL,
		Logger: log.Default(),
	}
	for _, f := range fn {
		f(opts)
	}

	// Validate options
	if opts.Token == "" {
		return nil, fmt.Errorf("missing telegram token")
	}
	if opts.WAQI == nil {
		return nil, fmt.Errorf("missing WAQI service instance")
	}

	allowedUserIDs := make(map[int]interface{})
	allowedUsernames := make(map[string]interface{})
	for _, s := range opts.AllowedUsernames {
		userID, err := strconv.Atoi(s)
		if err == nil {
			allowedUserIDs[userID] = nil
		} else {
			allowedUsernames[s] = nil
		}
	}
	if len(allowedUserIDs) == 0 && len(allowedUsernames) == 0 {
		return nil, fmt.Errorf("missing allowed usernames")
	}

	// Create database context
	db, err := NewDB(opts.DBPath, opts.Logger)
	if err != nil {
		return nil, err
	}

	// Create telegram Bot
	botSettings := telebot.Settings{
		URL:    opts.URL,
		Token:  opts.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	tgBot, err := telebot.NewBot(botSettings)
	if err != nil {
		opts.Logger.Printf("unable to connect to telegram: %s", err)
		db.Close()
		return nil, err
	}

	// Create Bot service
	bot := &botService{
		Bot:                tgBot,
		DB:                 db,
		AllowedUserIDs:     allowedUserIDs,
		AllowedUsernames:   allowedUsernames,
		WAQI:               opts.WAQI,
		SubscriptionsMutex: &sync.Mutex{},
		Subscriptions:      make(map[int]int),
		Screens:            &botScreens{tgBot, opts.Logger},
		Logger:             opts.Logger,
	}
	return bot, nil
}
