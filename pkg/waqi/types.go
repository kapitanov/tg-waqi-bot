package waqi

import (
	"encoding/json"
	"time"
)

// Level is an air quality level value
type Level string

const (
	// GoodLevel means air quality is considered satisfactory.
	// Maps to AQI from 0 to 50.
	GoodLevel Level = "good"

	// ModerateLevel means air quality is overall satisfactory
	// but there may be a moderate health concern for a very small number of people
	// who are unusually sensitive to air pollution
	// Maps to AQI from 51 to 100
	ModerateLevel Level = "moderate"

	// PossiblyUnhealthyLevel means members of sensitive groups may experience health effects.
	// Maps to AQI from 101 to 150.
	PossiblyUnhealthyLevel = "possibly_unhealthy"

	// UnhealthyLevel means everyone may begin to experience health effects
	// and members of sensitive groups may experience more serious health effects
	// Maps to AQI from 151 to 200.
	UnhealthyLevel Level = "unhealthy"

	// VeryUnhealthyLevel is means health warnings of emergency conditions.
	// Maps to AQI from 201 to 300.
	VeryUnhealthyLevel Level = "very_unhealthy"

	// HazardousLevel means a health alert.
	// Maps to AQI from 300 and above.
	HazardousLevel Level = "hazardous"
)

// String converts a value of Level into string
func (level Level) String() string {
	switch level {
	case GoodLevel:
		return "Good"
	case ModerateLevel:
		return "Satisfactory"
	case PossiblyUnhealthyLevel:
		return "Moderately polluted"
	case UnhealthyLevel:
		return "Poor"
	case VeryUnhealthyLevel:
		return "Very poor"
	case HazardousLevel:
		return "Hazardous"
	default:
		return string(level)
	}
}

// CalcAQILevel calculates an air quality level for raw AQI value
func CalcAQILevel(value float32) Level {
	if value < 51 {
		return GoodLevel
	}
	if value < 101 {
		return ModerateLevel
	}
	if value < 151 {
		return PossiblyUnhealthyLevel
	}
	if value < 201 {
		return UnhealthyLevel
	}
	if value < 300 {
		return VeryUnhealthyLevel
	}
	return HazardousLevel
}

// CalcPM10Level calculates a PM10 level for raw PM10 value
func CalcPM10Level(value float32) Level {
	if value < 51 {
		return GoodLevel
	}
	if value < 101 {
		return ModerateLevel
	}
	if value < 251 {
		return PossiblyUnhealthyLevel
	}
	if value < 351 {
		return UnhealthyLevel
	}
	if value < 430 {
		return VeryUnhealthyLevel
	}
	return HazardousLevel
}

// CalcPM25Level calculates a PM2.5 level for raw PM2.5 value
func CalcPM25Level(value float32) Level {
	if value < 31 {
		return GoodLevel
	}
	if value < 61 {
		return ModerateLevel
	}
	if value < 91 {
		return PossiblyUnhealthyLevel
	}
	if value < 121 {
		return UnhealthyLevel
	}
	if value < 250 {
		return VeryUnhealthyLevel
	}
	return HazardousLevel
}

// CalcNO2Level calculates a NO2 level for raw NO2 value
func CalcNO2Level(value float32) Level {
	if value < 41 {
		return GoodLevel
	}
	if value < 81 {
		return ModerateLevel
	}
	if value < 181 {
		return PossiblyUnhealthyLevel
	}
	if value < 281 {
		return UnhealthyLevel
	}
	if value < 401 {
		return VeryUnhealthyLevel
	}
	return HazardousLevel
}

// CalcO3Level calculates an O3 level for raw O3 value
func CalcO3Level(value float32) Level {
	if value < 51 {
		return GoodLevel
	}
	if value < 101 {
		return ModerateLevel
	}
	if value < 169 {
		return PossiblyUnhealthyLevel
	}
	if value < 209 {
		return UnhealthyLevel
	}
	if value < 748 {
		return VeryUnhealthyLevel
	}
	return HazardousLevel
}

// CalcCOLevel calculates a CO level for raw CO value
func CalcCOLevel(value float32) Level {
	if value < 1.1 {
		return GoodLevel
	}
	if value < 2.1 {
		return ModerateLevel
	}
	if value < 10 {
		return PossiblyUnhealthyLevel
	}
	if value < 17 {
		return UnhealthyLevel
	}
	if value < 34 {
		return VeryUnhealthyLevel
	}
	return HazardousLevel
}

// CalcSO2Level calculates a SO2 level for raw SO2 value
func CalcSO2Level(value float32) Level {
	if value < 41 {
		return GoodLevel
	}
	if value < 81 {
		return ModerateLevel
	}
	if value < 381 {
		return PossiblyUnhealthyLevel
	}
	if value < 801 {
		return UnhealthyLevel
	}
	if value < 1600 {
		return VeryUnhealthyLevel
	}
	return HazardousLevel
}

// Error represents a service-specific error
type Error string

// String converts an instance of Error into String
func (e Error) String() string {
	return string(e)
}

// Error retrieves an error message from an instance of Error
func (e Error) Error() string {
	return string(e)
}

// Status contains aggregated air quality status
type Status struct {
	Station *Station `json:"station"`

	// Measurement time
	Time time.Time `json:"time"`

	// Air quality index value
	AQI float32 `json:"aqi"`

	// Air quality index level
	Level Level `json:"level"`

	// Particulate matter 2.5 measurement
	PM25 *float32 `json:"pm25"`

	// Particulate matter 10 measurement
	PM10 *float32 `json:"pm10"`

	// Ozone measurement
	O3 *float32 `json:"o3"`

	// Nitrogen dioxide measurement
	NO2 *float32 `json:"no2"`

	// Sulfur dioxide measurement
	SO2 *float32 `json:"so2"`

	// Carbon monoxide level measurement
	CO *float32 `json:"co"`
}

// String converts Status to string
func (s *Status) String() string {
	str, _ := json.Marshal(s)
	return string(str)
}

// Equal checks two Status values for equality
func (s *Status) Equal(other *Status) bool {
	if s.Station.ID != other.Station.ID {
		return false
	}

	if s.Level != other.Level {
		return false
	}

	if !areValuesEqual(s.PM25, other.PM25) {
		return false
	}

	if !areValuesEqual(s.PM10, other.PM10) {
		return false
	}

	if !areValuesEqual(s.O3, other.O3) {
		return false
	}

	if !areValuesEqual(s.NO2, other.NO2) {
		return false
	}

	if !areValuesEqual(s.SO2, other.SO2) {
		return false
	}

	if !areValuesEqual(s.CO, other.CO) {
		return false
	}

	return true
}

func areValuesEqual(x, y *float32) bool {
	if x == nil && y == nil {
		return true
	}

	if x == nil || y == nil {
		return false
	}

	return *x == *y
}

// Station contains weather station information
type Station struct {
	// Unique ID for the city monitoring station.
	ID int `json:"id"`

	// Name of the monitoring station
	Name string `json:"name"`

	// URL of the monitoring station website
	URL string `json:"url"`

	// Latitude of the monitoring station
	Lon float32 `json:"lon"`

	// Longitude of the monitoring station
	Lat float32 `json:"lat"`
}

// Service is an entry point for WAQI service
type Service interface {
	// GetByCity fetches current measurements for city
	GetByCity(city string) (*Status, error)

	// GetByStation fetches current measurements for station
	GetByStation(stationID int) (*Status, error)

	// GetByGeo fetches current measurements for geo coordinates
	GetByGeo(lat, lon float32) (*Status, error)

	// Subscribe adds a listener to updates
	Subscribe(stationID int, listener Listener)

	// Unsubscribe removes a listener
	Unsubscribe(stationID int, listener Listener)

	// StartUpdates starts background data updates
	StartUpdates()

	// StopUpdates stops background data updates
	StopUpdates()

	// Close shuts down service
	Close() error
}

// Listener receives updates on air quality status
type Listener interface {
	// Update handles a weather data update
	// prevStatus will be nil on first update
	// and not nil - on subsequent ones
	Update(status *Status, prevStatus *Status) error
}
