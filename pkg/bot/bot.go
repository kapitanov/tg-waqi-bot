package bot

import (
	"fmt"
	"github.com/kapitanov/tg-waqi-bot/pkg/waqi"
	"gopkg.in/tucnak/telebot.v2"
	"sync"
)

type botService struct {
	Bot                *telebot.Bot
	DB                 DB
	WAQI               waqi.Service
	AllowedUserIDs     map[int]interface{}
	AllowedUsernames   map[string]interface{}
	SubscriptionsMutex *sync.Mutex
	Subscriptions      map[int]int
}

// Start starts Bot
func (s *botService) Start() error {
	log.Printf("running as %s", s.Bot.Me.Username)
	log.Printf("allowed users:")
	for key := range s.AllowedUserIDs {
		log.Printf("  - %d", key)
	}
	for key := range s.AllowedUsernames {
		log.Printf("  - @%s", key)
	}

	// Configure bot
	s.Bot.Handle("/start", s.onStart)
	s.Bot.Handle(telebot.OnLocation, s.onLocation)
	s.Bot.Handle(telebot.OnCallback, s.onCallback)

	go s.Bot.Start()

	// Restore subscriptions
	m, err := s.DB.GetSubscribedStationIDs()
	if err != nil {
		return err
	}
	s.SubscriptionsMutex.Lock()
	defer s.SubscriptionsMutex.Unlock()
	s.Subscriptions = m
	log.Printf("got %d active subscriptions", len(m))
	for stationID := range m {
		s.WAQI.Subscribe(stationID, s)
		log.Printf("subscribed to station #%d", stationID)
	}

	log.Printf("bot is up and running")
	return nil
}

// Close shuts down Bot
func (s *botService) Close() {
	if s.Bot != nil {
		s.Bot.Stop()
		s.Bot = nil
	}

	s.DB.Close()
}

// onStart handles "/start" command
func (s *botService) onStart(m *telebot.Message) {
	log.Printf("got message \"%s\" from %d @%s", m.Text, m.Sender.ID, m.Sender.Username)
	s.handle(m, m.Chat, m.Sender, s.onStartCore)
}

// onStartCore handles "/start" command (without error handling)
func (s *botService) onStartCore(arg interface{}, chat *chatEntity) error {
	m := arg.(*telebot.Message)

	chat.SetStateNotSubscribed()

	err := s.DB.Update(chat)
	if err != nil {
		return err
	}

	err = sendWelcomeScreen(s.Bot, m.Chat, nil)
	if err != nil {
		return err
	}

	return nil
}

// onLocation handles location message
func (s *botService) onLocation(m *telebot.Message) {
	log.Printf("got location (%0.3f. %0.3f) from %d @%s", m.Location.Lat, m.Location.Lng, m.Sender.ID, m.Sender.Username)
	s.handle(m, m.Chat, m.Sender, s.onLocationCore)
}

// onLocationCore handles location message (without error handling)
func (s *botService) onLocationCore(arg interface{}, _ *chatEntity) error {
	m := arg.(*telebot.Message)

	status, err := s.WAQI.GetByGeo(m.Location.Lat, m.Location.Lng)
	if err != nil {
		log.Printf("unable to query status for location (%f, %f)", m.Location.Lat, m.Location.Lng)
		return sendErrorScreen(s.Bot, m.Chat)
	}

	return sendLocationScreen(s.Bot, m.Chat, status, nil)
}

// onCallback handles callbacks
func (s *botService) onCallback(c *telebot.Callback) {
	log.Printf("got callback \"%s\" from %d @%s", c.Data, c.Sender.ID, c.Sender.Username)
	s.handle(c, c.Message.Chat, c.Sender, s.onCallbackCore)
}

// onCallbackCore handles callbacks
func (s *botService) onCallbackCore(arg interface{}, chat *chatEntity) error {
	c := arg.(*telebot.Callback)
	callback, err := parseCallbackJSON(c.Data)
	if err != nil {
		return fmt.Errorf("malformed callback data: \"%s\". %s", c.Data, err)
	}

	switch callback.Type {
	case callbackTypeSubscribe:
		err = s.onCallbackSubscribe(c, callback, c.Sender, chat)
		break
	case callbackTypeUnsubscribe:
		err = s.onCallbackUnsubscribe(c, callback, c.Sender, chat)
		break
	case callbackTypeRefresh:
		err = s.onCallbackRefresh(c, callback, c.Sender, chat)
		break
	default:
		err = fmt.Errorf("unknown callback data: \"%s\"", c.Data)
		break
	}
	if err != nil {
		return err
	}

	return nil
}

// onCallbackSubscribe handles "subscribe" callbacks
func (s *botService) onCallbackSubscribe(c *telebot.Callback, d *callbackJSON, to telebot.Recipient, chat *chatEntity) error {
	// Set current air quality for specified station
	// We expected that this value is currently cached
	status, err := s.WAQI.GetByStation(d.StationID)
	if err != nil {
		return err
	}

	// Store subscription into DB
	chat.SetStateSubscribed(status.Station.ID)
	err = s.DB.Update(chat)
	if err != nil {
		return err
	}

	// Add in-memory subscription
	s.SubscriptionsMutex.Lock()
	defer s.SubscriptionsMutex.Unlock()
	counter, exists := s.Subscriptions[d.StationID]
	if !exists {
		counter = 0
		s.WAQI.Subscribe(d.StationID, s)
		log.Printf("subscribed to station #%d", d.StationID)
	}
	s.Subscriptions[d.StationID] = counter + 1

	// Show notification
	err = sendSubscribedScreen(s.Bot, to, status, c.Message)
	if err != nil {
		return err
	}

	return nil
}

// onCallbackUnsubscribe handles "unsubscribe" callbacks
func (s *botService) onCallbackUnsubscribe(c *telebot.Callback, d *callbackJSON, to telebot.Recipient, chat *chatEntity) error {
	// Set current air quality for specified station
	// We expected that this value is currently cached
	status, err := s.WAQI.GetByStation(d.StationID)
	if err != nil {
		return err
	}

	// Store subscription into DB
	chat.SetStateNotSubscribed()
	err = s.DB.Update(chat)
	if err != nil {
		return err
	}

	// Remove in-memory subscription
	s.SubscriptionsMutex.Lock()
	defer s.SubscriptionsMutex.Unlock()
	counter, exists := s.Subscriptions[d.StationID]
	if exists && counter == 1 {
		delete(s.Subscriptions, d.StationID)
		s.WAQI.Unsubscribe(d.StationID, s)
		log.Printf("unsubscribed from station #%d", d.StationID)
	} else {
		s.Subscriptions[d.StationID] = counter - 1
	}

	// Show notification
	err = sendLocationScreen(s.Bot, to, status, c.Message)
	if err != nil {
		return err
	}

	return nil
}

// onCallbackRefresh handles "refresh" callbacks
func (s *botService) onCallbackRefresh(c *telebot.Callback, d *callbackJSON, to telebot.Recipient, chat *chatEntity) error {
	// Set current air quality for specified station
	status, err := s.WAQI.GetByStation(d.StationID)
	if err != nil {
		return err
	}

	// Show notification
	if chat.State == StateSubscribed && chat.SubscribedToStationID == d.StationID {
		err = sendSubscribedScreen(s.Bot, to, status, c.Message)
	} else {
		err = sendLocationScreen(s.Bot, to, status, c.Message)
	}
	if err != nil {
		return err
	}

	return nil
}

// handle implements unified telegram event handler (with error handling)
func (s *botService) handle(arg interface{}, c *telebot.Chat, u *telebot.User, f func(interface{}, *chatEntity) error) {
	err := s.handleCore(arg, c, u, f)
	if err != nil {
		switch s := arg.(type) {
		case *telebot.Message:
			log.Printf("failed to handle message %d from %d: %s", s.ID, c.ID, err)
			break
		case *telebot.Callback:
			log.Printf("failed to handle callback \"%s\" from %d: %s", s.ID, c.ID, err)
			break
		default:
			log.Printf("failed to handle message from %d: %s", c.ID, err)
			break
		}

		err = sendErrorScreen(s.Bot, c)
		if err != nil {
			log.Printf("error while sending ErrorScreen: %s", err)
		}
	}
}

// handleCore implements unified telegram event handler (without error handling)
func (s *botService) handleCore(arg interface{}, c *telebot.Chat, u *telebot.User, f func(interface{}, *chatEntity) error) error {
	_, userIDAllowed := s.AllowedUserIDs[u.ID]
	_, usernameAllowed := s.AllowedUsernames[u.Username]

	if !userIDAllowed && !usernameAllowed {
		return sendForbiddenScreen(s.Bot, c)
	}

	chat, err := s.DB.GetOrCreate(c.ID, u.ID, u.Username)
	if err != nil {
		return err
	}

	err = s.DB.Update(chat)
	if err != nil {
		return err
	}

	err = f(arg, chat)
	if err != nil {
		return err
	}

	return nil
}

// Update handles a weather data update
// prevStatus will be nil on first update
// and not nil - on subsequent ones
func (s *botService) Update(status *waqi.Status, prevStatus *waqi.Status) error {
	log.Printf("UPDATE: was %s, became %s", status, prevStatus)

	chats, err := s.DB.GetSubscribedChats(status.Station.ID)
	if err != nil {
		return err
	}

	for _, chat := range chats {
		err = sendUpdatedScreen(s.Bot, chat, status, prevStatus, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
