package waqi

import (
	"log"
	"strings"
	"time"
)

const (
	// DefaultURL is default WAQI service root URL
	DefaultURL = "https://api.waqi.info/"

	// DefaultCacheDuration is default WAQI service cache duration
	DefaultCacheDuration = 15 * time.Minute
)

type options struct {
	URL           string
	Token         string
	CachePath     string
	CacheDuration time.Duration
	Logger        *log.Logger
}

// Normalize normalizes options
func (opts *options) Normalize() {
	opts.URL = strings.TrimRight(opts.URL, "/")
}

// Option is a configuration option for NewServer function
type Option func(*options)

// URLOption sets root URL
func URLOption(url string) Option {
	return func(opts *options) {
		opts.URL = url
	}
}

// TokenOption sets access token
func TokenOption(token string) Option {
	return func(opts *options) {
		opts.Token = token
	}
}

// CachePathOption sets path to cache file
func CachePathOption(path string) Option {
	return func(opts *options) {
		opts.CachePath = path
	}
}

// CacheDurationOption sets max cache duration
func CacheDurationOption(duration time.Duration) Option {
	return func(opts *options) {
		opts.CacheDuration = duration
	}
}

// LoggerOption sets logger instance
func LoggerOption(logger *log.Logger) Option {
	return func(opts *options) {
		opts.Logger = logger
	}
}

// NewService creates new instance of Service
func NewService(fn ...Option) (Service, error) {
	opts := &options{
		URL:           DefaultURL,
		CacheDuration: DefaultCacheDuration,
		Logger:        log.Default(),
	}
	for _, f := range fn {
		f(opts)
	}

	opts.Normalize()

	adapter := newServiceAdapter(opts.URL, opts.Token, opts.Logger)
	if opts.CachePath != "" {
		var err error
		adapter, err = newCachingServiceAdapter(adapter, opts.CachePath, opts.CacheDuration)
		if err != nil {
			return nil, err
		}
	}

	s := &service{
		adapter: adapter,
		fetcher: newFetcher(adapter, opts.Logger),
	}
	return s, nil
}

type service struct {
	adapter adapter
	fetcher *fetcher
}

// GetByCity fetches current measurements for city
func (s *service) GetByCity(city string) (*Status, error) {
	return s.adapter.GetByCity(city)
}

// GetByStation fetches current measurements for station
func (s *service) GetByStation(stationID int) (*Status, error) {
	return s.adapter.GetByStation(stationID)
}

// GetByGeo fetches current measurements for geo coordinates
func (s *service) GetByGeo(lat, lon float32) (*Status, error) {
	return s.adapter.GetByGeo(lat, lon)
}

// Subscribe adds a listener to updates
func (s *service) Subscribe(stationID int, listener Listener) {
	s.fetcher.Subscribe(stationID, listener)
}

// Unsubscribe removes a listener
func (s *service) Unsubscribe(stationID int, listener Listener) {
	s.fetcher.Unsubscribe(stationID, listener)
}

// StartUpdates starts background data updates
func (s *service) StartUpdates() {
	s.fetcher.StartUpdates()
}

// StopUpdates stops background data updates
func (s *service) StopUpdates() {
	s.fetcher.StopUpdates()
}

// Close shuts down service
func (s *service) Close() error {
	return s.adapter.Close()
}
