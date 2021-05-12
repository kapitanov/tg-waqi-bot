package waqi

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"strings"
	"time"
)

type cachedStatus struct {
	time   time.Time
	status *Status
}

type cachingServiceAdapter struct {
	adapter adapter
	db      *leveldb.DB
	maxAge  time.Duration
}

func newCachingServiceAdapter(adapter adapter, path string, maxAge time.Duration) (adapter, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &cachingServiceAdapter{adapter, db, maxAge}, nil

}

// GetByCity fetches current measurements for city
func (s *cachingServiceAdapter) GetByCity(city string) (*Status, error) {
	return s.GetOrAdd(s.GetCityKey(city), func() (*Status, error) {
		return s.adapter.GetByCity(city)
	})
}

// GetByStation fetches current measurements for station
func (s *cachingServiceAdapter) GetByStation(stationID int) (*Status, error) {
	return s.GetOrAdd(s.GetStationKey(stationID), func() (*Status, error) {
		return s.adapter.GetByStation(stationID)
	})
}

// GetByGeo fetches current measurements for geo coordinates
func (s *cachingServiceAdapter) GetByGeo(lat, lon float32) (*Status, error) {
	return s.GetOrAdd(s.GetGeoKey(lat, lon), func() (*Status, error) {
		return s.adapter.GetByGeo(lat, lon)
	})
}

// GetOrAdd gets a value from cache or fetches a new one
func (s *cachingServiceAdapter) GetOrAdd(key string, fn func() (*Status, error)) (*Status, error) {
	raw, err := s.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return s.FetchAndPut(fn)
		}

		return nil, err
	}

	var cached cachedStatus
	err = json.Unmarshal(raw, &cached)
	if err != nil {
		return nil, err
	}

	age := time.Now().Sub(cached.time)
	if age.Milliseconds() >= s.maxAge.Milliseconds() {
		return s.FetchAndPut(fn)
	}

	return cached.status, nil
}

// FetchAndPut fetches a value and stores it into cache
func (s *cachingServiceAdapter) FetchAndPut(fn func() (*Status, error)) (*Status, error) {
	status, err := fn()
	if err != nil {
		return nil, err
	}

	cached := &cachedStatus{
		status: status,
		time:   time.Now(),
	}
	bytes, err := json.Marshal(cached)
	if err != nil {
		return nil, err
	}

	keys := s.GetKeys(status)
	for _, key := range keys {
		err = s.db.Put([]byte(key), bytes, nil)
		if err != nil {
			return nil, err
		}
	}

	return status, nil
}

// GetCityKey returns cache key for city
func (s *cachingServiceAdapter) GetCityKey(name string) string {
	return fmt.Sprintf("city/%s", strings.ToLower(name))
}

// GetStationKey returns cache key for station ID
func (s *cachingServiceAdapter) GetStationKey(stationID int) string {
	return fmt.Sprintf("station/%d", stationID)
}

// GetGeoKey returns cache key for geo coordinates
func (s *cachingServiceAdapter) GetGeoKey(lat, lon float32) string {
	return fmt.Sprintf("geo/%0.2f/%0.2f", lat, lon)
}

// GetKeys returns a slice containing all possible caching keys for specified status
func (s *cachingServiceAdapter) GetKeys(status *Status) []string {
	keys := []string{
		s.GetCityKey(status.Station.Name),
		s.GetStationKey(status.Station.ID),
		s.GetGeoKey(status.Station.Lat, status.Station.Lon),
	}
	return keys
}

// Close shuts down adapter
func (s *cachingServiceAdapter) Close() error {
	err := s.adapter.Close()
	if err != nil {
		return err
	}

	err = s.db.Close()
	if err != nil {
		return err
	}

	return nil
}
