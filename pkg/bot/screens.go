package bot

import (
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/kapitanov/tg-waqi-bot/pkg/waqi"
	"gopkg.in/tucnak/telebot.v2"
	"regexp"
	"strings"
	"time"
)

func sendForbiddenScreen(bot *telebot.Bot, to telebot.Recipient) error {
	text := fmt.Sprintf("%s Sorry, this bot is private. You are not in allowed user list.", emoji.NoEntry)

	markup := &telebot.ReplyMarkup{
		ReplyKeyboardRemove: true,
	}

	_, err := bot.Send(to, text, markup, telebot.ModeHTML)
	log.Printf("sent ForbiddenScreen to %s", to.Recipient())
	return err
}

func sendErrorScreen(bot *telebot.Bot, to telebot.Recipient) error {
	text := fmt.Sprintf("%s Error! Something went wrong on server side", emoji.ExclamationMark)

	markup := &telebot.ReplyMarkup{
		ReplyKeyboardRemove: true,
	}

	_, err := bot.Send(to, text, markup, telebot.ModeHTML)
	log.Printf("sent ErrorScreen to %s", to.Recipient())
	return err
}

func sendWelcomeScreen(bot *telebot.Bot, to telebot.Recipient, message telebot.Editable) error {
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

	return sendScreen("WelcomeScreen", bot, to, message, text, markup, telebot.ModeHTML)
}

func sendLocationScreen(bot *telebot.Bot, to telebot.Recipient, status *waqi.Status, message telebot.Editable) error {
	text := generateStatusScreen(status)

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
	return sendScreen(name, bot, to, message, text, markup, telebot.ModeHTML, telebot.NoPreview)
}

func sendSubscribedScreen(bot *telebot.Bot, to telebot.Recipient, status *waqi.Status, message telebot.Editable) error {
	text := generateStatusScreen(status)

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
	return sendScreen(name, bot, to, message, text, markup, telebot.ModeHTML, telebot.NoPreview)
}

func sendUpdatedScreen(bot *telebot.Bot, to telebot.Recipient, status *waqi.Status, prevStatus *waqi.Status, message telebot.Editable) error {
	if prevStatus == nil {
		return sendSubscribedScreen(bot, to, status, message)
	}

	text := generateDeltaStatusScreen(status, prevStatus)

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
	return sendScreen(name, bot, to, message, text, markup, telebot.ModeHTML, telebot.NoPreview)
}

func generateStatusScreen(status *waqi.Status) string {
	// First row - title and hyperlink
	text := ""
	stationName := status.Station.Name
	if status.Station.URL != "" {
		text += fmt.Sprintf("<b><a href=\"%s\">%s</a></b>\n\n", status.Station.URL, stationName)
	} else {
		text += fmt.Sprintf("<b>%s</b>\n\n", stationName)
	}

	// Second row - status icon and text
	text += fmt.Sprintf("Air quality: %s <code>%s</code>\n", getLevelIcon(status.Level), status.Level.String())
	text += "\n"

	// Third and subsequent rows - parameters
	unit := "μg/m3"
	text = appendStatusParameter(text, "AQI  ", &status.AQI, nil, "", waqi.CalcAQILevel)
	text = appendStatusParameter(text, "PM2.1", status.PM25, nil, unit, waqi.CalcPM25Level)
	text = appendStatusParameter(text, "PM10 ", status.PM10, nil, unit, waqi.CalcPM10Level)
	text = appendStatusParameter(text, "O3   ", status.O3, nil, unit, waqi.CalcO3Level)
	text = appendStatusParameter(text, "NO2  ", status.NO2, nil, unit, waqi.CalcNO2Level)
	text = appendStatusParameter(text, "SO2  ", status.SO2, nil, unit, waqi.CalcSO2Level)
	text = appendStatusParameter(text, "CO   ", status.CO, nil, unit, waqi.CalcCOLevel)

	// Last row - date and time
	text += fmt.Sprintf("\nUpdated at %s UTC", status.Time.Format("2006-Jan-2 15:04:05"))

	return text
}

func generateDeltaStatusScreen(status *waqi.Status, prevStatus *waqi.Status) string {
	// First row - title and hyperlink
	text := ""
	stationName := status.Station.Name
	if status.Station.URL != "" {
		text += fmt.Sprintf("<b><a href=\"%s\">%s</a></b>\n\n", status.Station.URL, stationName)
	} else {
		text += fmt.Sprintf("<b>%s</b>\n\n", stationName)
	}

	// Second row - status icon and text
	text += fmt.Sprintf("Air quality: %s <code>%s</code>\n", getLevelIcon(status.Level), status.Level.String())
	text += "\n"

	// Third and subsequent rows - parameters
	unit := "μg/m3"
	text = appendStatusParameter(text, "AQI  ", &status.AQI, &prevStatus.AQI, "", waqi.CalcAQILevel)
	text = appendStatusParameter(text, "PM2.1", status.PM25, prevStatus.PM25, unit, waqi.CalcPM25Level)
	text = appendStatusParameter(text, "PM10 ", status.PM10, prevStatus.PM10, unit, waqi.CalcPM10Level)
	text = appendStatusParameter(text, "O3   ", status.O3, prevStatus.O3, unit, waqi.CalcO3Level)
	text = appendStatusParameter(text, "NO2  ", status.NO2, prevStatus.NO2, unit, waqi.CalcNO2Level)
	text = appendStatusParameter(text, "SO2  ", status.SO2, prevStatus.SO2, unit, waqi.CalcSO2Level)
	text = appendStatusParameter(text, "CO   ", status.CO, prevStatus.CO, unit, waqi.CalcCOLevel)

	// Last row - date and time
	text += fmt.Sprintf("\nUpdated at %s UTC", status.Time.Format("2006-Jan-2 15:04:05"))
	
	return text
}

func getLevelIcon(level waqi.Level) string {
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

func appendStatusParameter(text, name string, value *float32, prevValue *float32, unit string, calcLevel func(float32) waqi.Level) string {
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
			iconStr = getLevelIcon(calcLevel(*value))
		}

		prevStr := ""
		if prevValue != nil {
			prevStr = fmt.Sprintf(" (was <code>%0.1f</code>)", *prevValue)
		}

		text += fmt.Sprintf("<code>%s: %s %s %s</code>%s\n", name, valueStr, unit, iconStr, prevStr)
	}
	return text
}

func sendScreen(name string, bot *telebot.Bot, to telebot.Recipient, message telebot.Editable, text string, options ...interface{}) error {
	var err error
	if message != nil {
		_, err = bot.Edit(message, text, options...)
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
		log.Printf("sent %s to %s updating %s from %d", name, to.Recipient(), msgID, chatID)
	} else {
		_, err = bot.Send(to, text, options...)
		log.Printf("sent %s to %s", name, to.Recipient())
	}

	return err
}
