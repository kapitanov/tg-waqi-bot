package bot_test

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kapitanov/tg-waqi-bot/pkg/bot"
)

func TestGetOrCreate(t *testing.T) {
	a := assert.New(t)
	dir, err := os.MkdirTemp(os.TempDir(), "*")
	a.Nil(err)

	filepath := path.Join(dir, "temp.dat")
	defer func() {
		_ = os.Remove(filepath)
	}()

	db, err := bot.NewDB(filepath, log.Default())
	a.Nil(err)
	defer db.Close()

	chatID := int64(1234)
	userID := 465
	username := "username"

	// At first call an entity should be created
	e1, err := db.GetOrCreate(chatID, userID, username)
	a.Nil(err)
	a.Equal(chatID, e1.ChatID)
	a.Equal(userID, e1.UserID)
	a.Equal(username, e1.UserName)
	a.Equal(bot.StateNotSubscribed, e1.State)
	a.Equal(0, e1.SubscribedToStationID)

	// At subsequent calls an entity should be fetched
	e2, err := db.GetOrCreate(chatID, userID, username)
	a.Nil(err)
	a.Equal(e1.ChatID, e2.ChatID)
	a.Equal(e1.UserID, e2.UserID)
	a.Equal(e1.UserName, e2.UserName)
	a.Equal(e1.State, e2.State)
	a.Equal(e1.SubscribedToStationID, e2.SubscribedToStationID)
	a.Equal(e1.Updated, e2.Updated)
}

func TestUpdate(t *testing.T) {
	a := assert.New(t)
	dir, err := os.MkdirTemp(os.TempDir(), "*")
	a.Nil(err)

	filepath := path.Join(dir, "temp.dat")
	defer func() {
		_ = os.Remove(filepath)
	}()

	db, err := bot.NewDB(filepath, log.Default())
	a.Nil(err)
	defer db.Close()

	chatID := int64(1234)
	userID := 465
	username := "username"

	// At first call an entity should be created
	e1, err := db.GetOrCreate(chatID, userID, username)
	a.Nil(err)
	a.Equal(chatID, e1.ChatID)
	a.Equal(userID, e1.UserID)
	a.Equal(username, e1.UserName)
	a.Equal(bot.StateNotSubscribed, e1.State)
	a.Equal(0, e1.SubscribedToStationID)

	// Then an existing entity is updated
	e1.SetStateSubscribed(123)
	err = db.Update(e1)
	a.Nil(err)

	// Updated values should be persisted
	e2, err := db.GetOrCreate(chatID, userID, username)
	a.Nil(err)
	a.Equal(e1.ChatID, e2.ChatID)
	a.Equal(e1.UserID, e2.UserID)
	a.Equal(e1.UserName, e2.UserName)
	a.Equal(e1.State, e2.State)
	a.Equal(e1.SubscribedToStationID, e2.SubscribedToStationID)
	a.Equal(e1.Updated, e2.Updated)
}

func TestGetSubscribedStationIDs(t *testing.T) {
	a := assert.New(t)
	dir, err := os.MkdirTemp(os.TempDir(), "*")
	a.Nil(err)

	filepath := path.Join(dir, "temp.dat")
	defer func() {
		_ = os.Remove(filepath)
	}()

	db, err := bot.NewDB(filepath, log.Default())
	a.Nil(err)
	defer db.Close()

	// Insert 1st entity
	chatID1 := int64(1234)
	userID1 := 465
	username1 := "username1"
	e1, err := db.GetOrCreate(chatID1, userID1, username1)
	a.Nil(err)
	e1.SetStateSubscribed(123)
	err = db.Update(e1)
	a.Nil(err)

	ids, err := db.GetSubscribedStationIDs()
	a.Nil(err)
	a.Len(ids, 1)
	a.Equal(1, ids[e1.SubscribedToStationID])

	// Insert 2nd entity
	chatID2 := int64(1235)
	userID2 := 466
	username2 := "username2"
	e2, err := db.GetOrCreate(chatID2, userID2, username2)
	a.Nil(err)
	e2.SetStateSubscribed(124)
	err = db.Update(e2)
	a.Nil(err)

	ids, err = db.GetSubscribedStationIDs()
	a.Nil(err)
	a.Len(ids, 2)
	a.Equal(1, ids[e1.SubscribedToStationID])
	a.Equal(1, ids[e2.SubscribedToStationID])

	// Update 2nd entity
	e2.SetStateNotSubscribed()
	err = db.Update(e2)
	a.Nil(err)

	ids, err = db.GetSubscribedStationIDs()
	a.Nil(err)
	a.Len(ids, 1)
	a.Equal(1, ids[e1.SubscribedToStationID])
}

func TestGetSubscribedChats(t *testing.T) {
	a := assert.New(t)
	dir, err := os.MkdirTemp(os.TempDir(), "*")
	a.Nil(err)

	filepath := path.Join(dir, "temp.dat")
	defer func() {
		_ = os.Remove(filepath)
	}()

	db, err := bot.NewDB(filepath, log.Default())
	a.Nil(err)
	defer db.Close()

	// Insert 1st entity
	chatID1 := int64(1234)
	userID1 := 465
	username1 := "username1"
	e1, err := db.GetOrCreate(chatID1, userID1, username1)
	a.Nil(err)
	e1.SetStateSubscribed(123)
	err = db.Update(e1)
	a.Nil(err)

	chats, err := db.GetSubscribedChats(e1.SubscribedToStationID)
	a.Nil(err)
	a.Len(chats, 1)
	a.Equal(e1.ChatID, chats[0].ChatID)

	// Insert 2nd entity
	chatID2 := int64(1235)
	userID2 := 466
	username2 := "username2"
	e2, err := db.GetOrCreate(chatID2, userID2, username2)
	a.Nil(err)
	e2.SetStateSubscribed(123)
	err = db.Update(e2)
	a.Nil(err)
	e1.SetStateSubscribed(123)
	err = db.Update(e1)
	a.Nil(err)

	chats, err = db.GetSubscribedChats(123)
	a.Nil(err)
	a.Len(chats, 2)
	a.Equal(e1.ChatID, chats[0].ChatID)
	a.Equal(e2.ChatID, chats[1].ChatID)

	// Update 2nd entity
	e2.SetStateNotSubscribed()
	err = db.Update(e2)
	a.Nil(err)
	e1.SetStateSubscribed(124)
	err = db.Update(e1)
	a.Nil(err)

	chats, err = db.GetSubscribedChats(123)
	a.Nil(err)
	a.Len(chats, 0)

	chats, err = db.GetSubscribedChats(124)
	a.Nil(err)
	a.Len(chats, 1)
	a.Equal(e1.ChatID, chats[0].ChatID)
}
