package main

import (
	"github.com/gorilla/sessions"
	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	begin int = iota
	reservationData
	reservationTime
	menu
	location
	confirmation
	end
)

func (client client) reply(event *linebot.Event) error {
	prevStep := client.session.Values["step"]
	switch prevStep {
	case "", begin:
		if err := replyReservationDate(event, client.bot, client.session); err != nil {
			return err
		}
	case reservationData:
		if err := replyReservationTime(event, client.bot, client.session); err != nil {
			return err
		}
	case reservationTime:
		if err := replyMenu(event, client.bot, client.session); err != nil {
			return err
		}
	case menu:
		if err := replyLocation(event, client.bot, client.session); err != nil {
			return err
		}
	case location:
		if err := replyConfirmation(event, client.bot, client.session); err != nil {
			return err
		}
	case confirmation, end:
		if err := replyThankYou(event, client.bot, client.session); err != nil {
			return err
		}
	default:
		if err := replySorry(event, client.bot, client.session); err != nil {
			return err
		}
	}
	return nil
}

func replyReservationDate(event *linebot.Event, bot *linebot.Client, session *sessions.Session) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeReservationDateMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyReservationTime(event *linebot.Event, bot *linebot.Client, session *sessions.Session) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeReservationTimeMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyMenu(event *linebot.Event, bot *linebot.Client, session *sessions.Session) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeMenuTextMessage(), makeMenuMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyLocation(event *linebot.Event, bot *linebot.Client, session *sessions.Session) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeLocationMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyConfirmation(event *linebot.Event, bot *linebot.Client, session *sessions.Session) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeConfirmationTextMessage(), makeConfirmationButtonMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyThankYou(event *linebot.Event, bot *linebot.Client, session *sessions.Session) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeThankYouMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replySorry(event *linebot.Event, bot *linebot.Client, session *sessions.Session) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeSorryMessage(), makeReservationDateMessage()).Do(); err != nil {
		return err
	}
	return nil
}
