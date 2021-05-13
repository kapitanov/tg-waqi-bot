package bot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const (
	StateNotSubscribed = "not_subscribed"
	StateSubscribed    = "subscribed"
)

type chatEntity struct {
	ChatID                int64     `gorm:"column:id;unique_index;primary_key"`
	UserID                int       `gorm:"column:user_id;unique_index"`
	UserName              string    `gorm:"column:user_name"`
	State                 string    `gorm:"column:state;index"`
	SubscribedToStationID int       `gorm:"column:station_id;index"`
	Updated               time.Time `gorm:"column:updated"`
}

// TableName overrides the table name for chatEntity
func (chatEntity) TableName() string {
	return "chats"
}

// Recipient returns legit Telegram chat_id or username
func (e chatEntity) Recipient() string {
	return fmt.Sprintf("%d", e.ChatID)
}

// SetStateNotSubscribed moves entity into "not_subscribed" state
func (e *chatEntity) SetStateNotSubscribed() {
	e.State = StateNotSubscribed
	e.SubscribedToStationID = 0
}

// SetStateSubscribed moves entity into "subscribed" state
func (e *chatEntity) SetStateSubscribed(stationID int) {
	e.State = StateSubscribed
	e.SubscribedToStationID = stationID
}

type DB interface {
	// GetOrCreate fetches a chat state from DB
	// If chat is not registered yet, it will be created
	GetOrCreate(chatID int64, userID int, username string) (*chatEntity, error)

	// Update stores chat state into DB
	Update(chat *chatEntity) error

	// GetSubscribedStationIDs returns map of stations with subscription
	// Map key is station ID and value is count of active subscriptions
	GetSubscribedStationIDs() (map[int]int, error)

	// GetSubscribedChats returns map of chats subscribed to specified station
	GetSubscribedChats(stationID int) ([]*chatEntity, error)

	// Close shuts down DB
	Close()
}

type database struct {
	context *gorm.DB
}

// NewDB creates new instance of DB
func NewDB(filepath string, logger *log.Logger) (DB, error) {
	dir := path.Dir(filepath)
	err := os.MkdirAll(dir, 0)
	if err != nil {
		logger.Printf("unable to create directory \"%s\": %v", dir, err)
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{
		Logger: gormLogger.New(logger, gormLogger.Config{
			LogLevel:                  gormLogger.Error,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		logger.Printf("unable to open database \"%s\": %v", filepath, err)
		return nil, err
	}

	err = db.AutoMigrate(&chatEntity{})
	if err != nil {
		logger.Printf("unable to migrate database \"%s\": %v", filepath, err)
		return nil, err
	}

	return &database{context: db}, nil
}

// GetOrCreate fetches a chat state from DB
// If chat is not registered yet, it will be created
func (db *database) GetOrCreate(chatID int64, userID int, username string) (*chatEntity, error) {
	var e chatEntity
	result := db.context.Where("id = ?", chatID).First(&e)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}

		e = chatEntity{
			ChatID:                chatID,
			UserID:                userID,
			UserName:              username,
			State:                 StateNotSubscribed,
			SubscribedToStationID: 0,
			Updated:               time.Now().UTC(),
		}
		result = db.context.Create(&e)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &e, nil
}

// Update stores chat state into DB
func (db *database) Update(chat *chatEntity) error {
	chat.Updated = time.Now().UTC()
	upd := map[string]interface{}{
		"user_name":  chat.UserName,
		"state":      chat.State,
		"station_id": chat.SubscribedToStationID,
		"updated":    chat.Updated,
	}
	result := db.context.Model(chat).Updates(upd)
	return result.Error
}

// GetSubscribedStationIDs returns map of stations with subscription
// Map key is station ID and value is count of active subscriptions
func (db *database) GetSubscribedStationIDs() (map[int]int, error) {
	sqlQuery :=
		"SELECT station_id, count(*) as count FROM chats\n" +
			"WHERE state = ?\n" +
			"GROUP BY station_id\n" +
			"ORDER BY station_id"
	var entities []struct {
		StationID int `gorm:"column:station_id"`
		Count     int `gorm:"column:count"`
	}
	err := db.context.Raw(sqlQuery, StateSubscribed).Scan(&entities).Error
	if err != nil {
		return nil, err
	}

	m := make(map[int]int)
	for _, e := range entities {
		m[e.StationID] = e.Count
	}

	return m, nil
}

// GetSubscribedChats returns map of chats subscribed to specified station
func (db *database) GetSubscribedChats(stationID int) ([]*chatEntity, error) {
	var entities []*chatEntity
	err := db.context.Model(&chatEntity{}).
		Where("state = ? AND station_id = ?", StateSubscribed, stationID).
		Scan(&entities).Error
	if err != nil {
		return nil, err
	}

	return entities, nil
}

// Close shuts down DB
func (db *database) Close() {
}
