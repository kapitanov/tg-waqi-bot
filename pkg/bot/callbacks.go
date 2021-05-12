package bot

import "encoding/json"

type callbackType string

const (
	callbackTypeSubscribe   callbackType = "subscribe"
	callbackTypeRefresh     callbackType = "refresh"
	callbackTypeUnsubscribe callbackType = "unsubscribe"
)

type callbackJSON struct {
	Type      callbackType `json:"type"`
	StationID int          `json:"station_id"`
	UID       string       `json:"uid"`
}

func (c callbackJSON) String() string {
	bytes, _ := json.Marshal(c)
	return string(bytes)
}

func parseCallbackJSON(raw string) (*callbackJSON, error) {
	var c callbackJSON
	err := json.Unmarshal([]byte(raw), &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
