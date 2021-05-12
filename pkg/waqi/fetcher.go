package waqi

import (
	pkgLog "github.com/kapitanov/tg-waqi-bot/pkg/log"
	"sync"
	"time"
)

var log = pkgLog.New("waqi")

type fetcher struct {
	adapter           adapter
	listeners         map[int]*stationFetcher
	mutex             *sync.Mutex
	isRunning         bool
	areUpdatesRunning bool
	sleepDuration     time.Duration
	ticker            *time.Ticker
	done              chan bool
}

func newFetcher(adapter adapter) *fetcher {
	f := &fetcher{
		adapter:       adapter,
		listeners:     make(map[int]*stationFetcher),
		mutex:         &sync.Mutex{},
		sleepDuration: 10 * time.Minute,
		done:          make(chan bool),
	}

	return f
}

// Subscribe adds a listener to updates
func (f *fetcher) Subscribe(stationID int, listener Listener) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	listeners, exists := f.listeners[stationID]
	if !exists {
		listeners = newStationFetcher(f.adapter, stationID)
		f.listeners[stationID] = listeners

		log.Printf("subscribed to station #%d", stationID)
	}

	listeners.Subscribe(listener)
}

// Unsubscribe removes a listener
func (f *fetcher) Unsubscribe(stationID int, listener Listener) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	listeners, exists := f.listeners[stationID]
	if !exists {
		return
	}

	listeners.Unsubscribe(listener)
	if listeners.Empty() {
		delete(f.listeners, stationID)

		log.Printf("unsubscribed from station #%d", stationID)
	}
}

// StartUpdates starts background data updates
func (f *fetcher) StartUpdates() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.ticker == nil {
		f.ticker = time.NewTicker(f.sleepDuration)
		log.Printf("starting background updates with period of %s", f.sleepDuration)
		go f.UpdateLoop()
	}
}

// StopUpdates stops background data updates
func (f *fetcher) StopUpdates() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.ticker != nil {
		f.ticker.Stop()
		f.ticker = nil

		f.done <- true
	}
}

// UpdateLoop runs background update loop
func (f *fetcher) UpdateLoop() {
	for {
		select {
		case <-f.done:
			return
		case <-f.ticker.C:
			f.UpdateOnce()
		}
	}
}

// UpdateOnce runs a single update
func (f *fetcher) UpdateOnce() {
	subscriptions := f.GetCurrentListeners()
	for _, listeners := range subscriptions {
		listeners.Update()
	}
}

// GetCurrentListeners returns a current set of listeners
func (f *fetcher) GetCurrentListeners() map[int]*stationFetcher {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	copyOfMap := make(map[int]*stationFetcher)
	for key, value := range f.listeners {
		copyOfMap[key] = value
	}

	return copyOfMap
}

type stationFetcher struct {
	adapter    adapter
	stationID  int
	listeners  []Listener
	mutex      *sync.Mutex
	prevStatus *Status
}

func newStationFetcher(adapter adapter, stationID int) *stationFetcher {
	f := &stationFetcher{
		adapter:   adapter,
		stationID: stationID,
		listeners: make([]Listener, 0),
		mutex:     &sync.Mutex{},
	}

	f.Update()

	return f
}

// Empty returns true is there are no listeners
func (f *stationFetcher) Empty() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return len(f.listeners) == 0
}

// Subscribe adds a listener
func (f *stationFetcher) Subscribe(listener Listener) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.listeners = append(f.listeners, listener)
}

// Unsubscribe removes a listener
func (f *stationFetcher) Unsubscribe(listener Listener) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for i := range f.listeners {
		if f.listeners[i] == listener {
			for j := i; j < len(f.listeners); j++ {
				f.listeners[j-1] = f.listeners[j]
			}

			f.listeners = f.listeners[0 : len(f.listeners)-1]
		}
	}
}

// Update fetches new value and pushes it to listeners
func (f *stationFetcher) Update() {
	status, err := f.adapter.GetByStation(f.stationID)
	if err != nil {
		log.Printf("unable to get data for station #%d: %s", f.stationID, err)
		return
	}

	prevStatus := f.prevStatus
	f.prevStatus = status

	if prevStatus == nil || prevStatus.Equal(status) {
		return
	}

	log.Printf("data for station #%d has been updated", f.stationID)
	f.PushToListeners(status, prevStatus)
}

// PushToListeners pushes new status to listeners
func (f *stationFetcher) PushToListeners(newStatus, prevStatus *Status) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for _, listener := range f.listeners {
		err := listener.Update(newStatus, prevStatus)
		if err != nil {
			log.Printf("unable to push data for station #%d to listener: %s", f.stationID, err)
		}
	}
}
