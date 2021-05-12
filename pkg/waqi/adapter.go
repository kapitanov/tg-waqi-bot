package waqi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// adapter is an adapter for WAQI service
type adapter interface {
	// GetByCity fetches current measurements for city
	GetByCity(city string) (*Status, error)

	// GetByStation fetches current measurements for station
	GetByStation(stationID int) (*Status, error)

	// GetByGeo fetches current measurements for geo coordinates
	GetByGeo(lat, lon float32) (*Status, error)

	// Close shuts down adapter
	Close() error
}

type serviceAdapter struct {
	url   string
	token string
}

func newServiceAdapter(url, token string) adapter {
	return &serviceAdapter{url, token}
}

// GetByCity fetches current measurements for city
func (s *serviceAdapter) GetByCity(city string) (*Status, error) {
	path := fmt.Sprintf("feed/%s/", city)
	return s.Get(path)
}

// GetByStation fetches current measurements for station
func (s *serviceAdapter) GetByStation(stationID int) (*Status, error) {
	path := fmt.Sprintf("feed/@%d/", stationID)
	return s.Get(path)
}

// GetByGeo fetches current measurements for geo coordinates
func (s *serviceAdapter) GetByGeo(lat, lon float32) (*Status, error) {
	path := fmt.Sprintf("feed/geo:%f;%f/", lat, lon)
	return s.Get(path)
}

// Get fetches current measurements by a relative URL
func (s *serviceAdapter) Get(path string) (*Status, error) {
	u := fmt.Sprintf("%s/%s?token=%s", s.url, path, url.QueryEscape(s.token))
	resp, err := http.Get(u)
	if err != nil {
		log.Printf("GET %s failed: %s", u, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("GET %s -> %d", u, resp.StatusCode)
		return nil, Error("server returned non-successful response")
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GET %s failed: %s", u, err)
		return nil, err
	}

	var raw responseJSON
	err = json.Unmarshal(buffer, &raw)
	if err != nil {
		log.Printf("GET %s failed: %s", u, err)
		return nil, err
	}

	if raw.Status == responseStatusError {
		log.Printf("GET %s -> %d: %s", u, resp.StatusCode, raw.Message)
		return nil, Error(fmt.Sprintf("server error: %s", raw.Message))
	}

	return raw.ToStatus(), nil
}

// Close shuts down adapter
func (s *serviceAdapter) Close() error {
	return nil
}
