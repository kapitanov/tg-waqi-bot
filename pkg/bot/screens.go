package bot

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/kapitanov/tg-waqi-bot/pkg/waqi"
)

type botScreens struct {
	bot    *telebot.Bot
	logger *log.Logger
}

func (s *botScreens) ForbiddenScreen(to telebot.Recipient) error {
	text := fmt.Sprintf("%s Sorry, this bot is private. You are not in allowed user list.", emoji.NoEntry)

	markup := &telebot.ReplyMarkup{
		ReplyKeyboardRemove: true,
	}

	_, err := s.bot.Send(to, text, markup, telebot.ModeHTML)
	s.logger.Printf("sent ForbiddenScreen to %s", to.Recipient())
	return err
}

func (s *botScreens) ErrorScreen(to telebot.Recipient) error {
	text := fmt.Sprintf("%s Error! Something went wrong on server side", emoji.ExclamationMark)

	markup := &telebot.ReplyMarkup{
		ReplyKeyboardRemove: true,
	}

	_, err := s.bot.Send(to, text, markup, telebot.ModeHTML)
	s.logger.Printf("sent ErrorScreen to %s", to.Recipient())
	return err
}

func (s *botScreens) WelcomeScreen(to telebot.Recipient, message telebot.Editable) error {
	text := fmt.Sprintf(
		"%s This Bot helps you track air quality at any location.\nSend me a location to get its current air quality index.",
		emoji.Umbrella)

	markup := &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{
				{
					Text:     "Send a location",
					Location: true,
				},
			},
		},
		OneTimeKeyboard: true,
	}

	return s.sendScreen("WelcomeScreen", to, message, text, markup, telebot.ModeHTML)
}

func (s *botScreens) LocationScreen(to telebot.Recipient, status *waqi.Status, message telebot.Editable) error {
	text := s.generateStatusScreen(status)

	uid := fmt.Sprintf("%d", time.Now().UTC().Unix())
	markup := &telebot.ReplyMarkup{
		ReplyKeyboardRemove: true,
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{
					Text: fmt.Sprintf("%s Refresh", emoji.CounterclockwiseArrowsButton),
					Data: callbackJSON{Type: callbackTypeRefresh, StationID: status.Station.ID, UID: uid}.String(),
				},
				{
					Text: fmt.Sprintf("%s Subscribe", emoji.CheckMarkButton),
					Data: callbackJSON{Type: callbackTypeSubscribe, StationID: status.Station.ID, UID: uid}.String(),
				},
			},
		},
		OneTimeKeyboard: true,
	}

	name := fmt.Sprintf("LocationScreen(%d)", status.Station.ID)
	return s.sendScreen(name, to, message, text, markup, telebot.ModeHTML, telebot.NoPreview)
}

func (s *botScreens) SubscribedScreen(to telebot.Recipient, status *waqi.Status, message telebot.Editable) error {
	text := s.generateStatusScreen(status)

	uid := fmt.Sprintf("%d", time.Now().UTC().Unix())
	markup := &telebot.ReplyMarkup{
		ReplyKeyboardRemove: true,
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{
					Text: fmt.Sprintf("%s Refresh", emoji.CounterclockwiseArrowsButton),
					Data: callbackJSON{Type: callbackTypeRefresh, StationID: status.Station.ID, UID: uid}.String(),
				},
				{
					Text: fmt.Sprintf("%s Unsubscribe", emoji.CrossMarkButton),
					Data: callbackJSON{Type: callbackTypeUnsubscribe, StationID: status.Station.ID, UID: uid}.String(),
				},
			},
		},
		OneTimeKeyboard: true,
	}

	name := fmt.Sprintf("SubscribedScreen(%d)", status.Station.ID)
	return s.sendScreen(name, to, message, text, markup, telebot.ModeHTML, telebot.NoPreview)
}

func (s *botScreens) UpdatedScreen(to telebot.Recipient, status *waqi.Status, prevStatus *waqi.Status, message telebot.Editable) error {
	if prevStatus == nil {
		return s.SubscribedScreen(to, status, message)
	}

	text := s.generateDeltaStatusScreen(status, prevStatus)

	uid := fmt.Sprintf("%d", time.Now().UTC().Unix())
	markup := &telebot.ReplyMarkup{
		ReplyKeyboardRemove: true,
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{
					Text: fmt.Sprintf("%s Refresh", emoji.CounterclockwiseArrowsButton),
					Data: callbackJSON{Type: callbackTypeRefresh, StationID: status.Station.ID, UID: uid}.String(),
				},
				{
					Text: fmt.Sprintf("%s Unsubscribe", emoji.CrossMarkButton),
					Data: callbackJSON{Type: callbackTypeUnsubscribe, StationID: status.Station.ID, UID: uid}.String(),
				},
			},
		},
		OneTimeKeyboard: true,
	}

	name := fmt.Sprintf("UpdatedScreen(%d)", status.Station.ID)
	return s.sendScreen(name, to, message, text, markup, telebot.ModeHTML, telebot.NoPreview)
}

func (s *botScreens) generateStatusScreen(status *waqi.Status) string {
	// First row - title and hyperlink
	text := ""
	stationName := status.Station.Name
	if status.Station.URL != "" {
		text += fmt.Sprintf("<b><a href=\"%s\">%s</a></b>\n\n", status.Station.URL, stationName)
	} else {
		text += fmt.Sprintf("<b>%s</b>\n\n", stationName)
	}

	// Second row - status icon and text
	text += fmt.Sprintf("Air quality: %s <code>%s</code>\n", s.getLevelIcon(status.Level), status.Level.String())
	text += "\n"

	// Third and subsequent rows - parameters
	unit := "μg/m3"
	text = s.appendStatusParameter(text, "AQI  ", &status.AQI, nil, "", waqi.CalcAQILevel)
	text = s.appendStatusParameter(text, "PM2.1", status.PM25, nil, unit, waqi.CalcPM25Level)
	text = s.appendStatusParameter(text, "PM10 ", status.PM10, nil, unit, waqi.CalcPM10Level)
	text = s.appendStatusParameter(text, "O3   ", status.O3, nil, unit, waqi.CalcO3Level)
	text = s.appendStatusParameter(text, "NO2  ", status.NO2, nil, unit, waqi.CalcNO2Level)
	text = s.appendStatusParameter(text, "SO2  ", status.SO2, nil, unit, waqi.CalcSO2Level)
	text = s.appendStatusParameter(text, "CO   ", status.CO, nil, unit, waqi.CalcCOLevel)

	// Last row - date and time
	text += fmt.Sprintf("\nUpdated at %s UTC", status.Time.Format("2006-Jan-2 15:04:05"))

	return text
}

func (s *botScreens) generateDeltaStatusScreen(status *waqi.Status, prevStatus *waqi.Status) string {
	// First row - title and hyperlink
	text := ""
	stationName := status.Station.Name
	if status.Station.URL != "" {
		text += fmt.Sprintf("<b><a href=\"%s\">%s</a></b>\n\n", status.Station.URL, stationName)
	} else {
		text += fmt.Sprintf("<b>%s</b>\n\n", stationName)
	}

	// Second row - status icon and text
	text += fmt.Sprintf("Air quality: %s <code>%s</code>\n", s.getLevelIcon(status.Level), status.Level.String())
	text += "\n"

	// Third and subsequent rows - parameters
	unit := "μg/m3"
	text = s.appendStatusParameter(text, "AQI  ", &status.AQI, &prevStatus.AQI, "", waqi.CalcAQILevel)
	text = s.appendStatusParameter(text, "PM2.1", status.PM25, prevStatus.PM25, unit, waqi.CalcPM25Level)
	text = s.appendStatusParameter(text, "PM10 ", status.PM10, prevStatus.PM10, unit, waqi.CalcPM10Level)
	text = s.appendStatusParameter(text, "O3   ", status.O3, prevStatus.O3, unit, waqi.CalcO3Level)
	text = s.appendStatusParameter(text, "NO2  ", status.NO2, prevStatus.NO2, unit, waqi.CalcNO2Level)
	text = s.appendStatusParameter(text, "SO2  ", status.SO2, prevStatus.SO2, unit, waqi.CalcSO2Level)
	text = s.appendStatusParameter(text, "CO   ", status.CO, prevStatus.CO, unit, waqi.CalcCOLevel)

	// Last row - date and time
	text += fmt.Sprintf("\nUpdated at %s UTC", status.Time.Format("2006-Jan-2 15:04:05"))

	return text
}

func (s *botScreens) getLevelIcon(level waqi.Level) string {
	var icon emoji.Emoji = ""
	switch level {
	case waqi.GoodLevel:
		icon = emoji.GreenSquare
		break
	case waqi.ModerateLevel:
		icon = emoji.YellowSquare
		break
	case waqi.PossiblyUnhealthyLevel:
		icon = emoji.OrangeSquare
		break
	case waqi.UnhealthyLevel:
		icon = emoji.RedSquare
		break
	case waqi.VeryUnhealthyLevel:
		icon = emoji.PurpleSquare
		break
	case waqi.HazardousLevel:
		icon = emoji.BrownSquare
		break
	}

	return icon.String()
}

func (s *botScreens) appendStatusParameter(text, name string, value *float32, prevValue *float32, unit string, calcLevel func(float32) waqi.Level) string {
	if value != nil {
		valueStr := fmt.Sprintf("%0.1f", *value)
		const minValueLength = 6
		if len(valueStr) < minValueLength {
			valueStr = strings.Repeat(" ", minValueLength-len(valueStr)) + valueStr
		}

		const minUnitLength = 5
		if len(unit) < minUnitLength {
			unit += strings.Repeat(" ", minUnitLength-len(unit))
		}

		iconStr := ""
		if calcLevel != nil {
			iconStr = s.getLevelIcon(calcLevel(*value))
		}

		prevStr := ""
		if prevValue != nil {
			prevStr = fmt.Sprintf(" (was <code>%0.1f</code>)", *prevValue)
		}

		text += fmt.Sprintf("<code>%s: %s %s %s</code>%s\n", name, valueStr, unit, iconStr, prevStr)
	}
	return text
}

func (s *botScreens) sendScreen(name string, to telebot.Recipient, message telebot.Editable, text string, options ...interface{}) error {
	var err error
	if message != nil {
		_, err = s.bot.Edit(message, text, options...)
		if err != nil {
			// Suppress errors like:
			//   Bad Request: message is not modified: specified new message content and reply markup are exactly
			//   the same as a current content and reply markup of the message (400)
			ok, _ := regexp.MatchString("message is not modified", err.Error())
			if ok {
				err = nil
			}
		}
		msgID, chatID := message.MessageSig()
		s.logger.Printf("sent %s to %s updating %s from %d", name, to.Recipient(), msgID, chatID)
	} else {
		_, err = s.bot.Send(to, text, options...)
		s.logger.Printf("sent %s to %s", name, to.Recipient())
	}

	return err
}
