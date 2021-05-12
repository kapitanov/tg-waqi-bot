package waqi

import "time"

const (
	responseStatusOK    = "ok"
	responseStatusError = "error"
)

// responseJSON is a root model for JSON response
type responseJSON struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Data    *dataJSON `json:"data"`
}

// dataJSON is a model for "data" node in WAQI response
type dataJSON struct {
	AQI  float32   `json:"aqi"`
	ID   int       `json:"idx"`
	City *cityJSON `json:"city"`
	IAQI *iaqiJSON `json:"iaqi"`
	Time *timeJSON `json:"time"`
}

// cityJSON is a model for "data.city" node in WAQI response
type cityJSON struct {
	Geo  []float32 `json:"geo"`
	Name string    `json:"name"`
	URL  string    `json:"url"`
}

// iaqiJSON is a model for "data.iaqi" node in WAQI response
type iaqiJSON struct {
	CO   *valueJSON `json:"co"`
	Dew  *valueJSON `json:"dew"`
	H    *valueJSON `json:"h"`
	NO2  *valueJSON `json:"no2"`
	O3   *valueJSON `json:"o3"`
	P    *valueJSON `json:"p"`
	PM10 *valueJSON `json:"pm10"`
	PM25 *valueJSON `json:"pm25"`
	SO2  *valueJSON `json:"so2"`
	T    *valueJSON `json:"t"`
	W    *valueJSON `json:"w"`
	WG   *valueJSON `json:"wg"`
}

// valueJSON is a model for "data.*.v" node in WAQI response
type valueJSON struct {
	Value *float32 `json:"v"`
}

// timeJSON is a model for "data.time" node in WAQI response
type timeJSON struct {
	ISO *time.Time `json:"iso"`
}

// ToStatus converts an API response into internal object
func (r responseJSON) ToStatus() *Status {
	var t time.Time
	if r.Data.Time != nil && r.Data.Time.ISO != nil {
		t = (*r.Data.Time.ISO).UTC()
	} else {
		t = time.Now().UTC()
	}
	status := &Status{
		Station: &Station{
			ID:   r.Data.ID,
			Name: r.Data.City.Name,
			URL:  r.Data.City.URL,
			Lon:  r.Data.City.Geo[0],
			Lat:  r.Data.City.Geo[1],
		},
		Time:  t,
		AQI:   r.Data.AQI,
		Level: CalcAQILevel(r.Data.AQI),
		PM25:  extractValueFromJSON(r.Data.IAQI.PM25),
		PM10:  extractValueFromJSON(r.Data.IAQI.PM10),
		O3:    extractValueFromJSON(r.Data.IAQI.O3),
		NO2:   extractValueFromJSON(r.Data.IAQI.NO2),
		SO2:   extractValueFromJSON(r.Data.IAQI.SO2),
		CO:    extractValueFromJSON(r.Data.IAQI.CO),
	}
	return status
}

func extractValueFromJSON(value *valueJSON) *float32 {
	if value == nil {
		return nil
	}

	return value.Value
}
