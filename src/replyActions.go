package main

import (
	"github.com/line/line-bot-sdk-go/linebot"
)

func replyReservationDate(event *linebot.Event, bot *linebot.Client) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeReservationDateMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyReservationTime(event *linebot.Event, bot *linebot.Client) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeReservationDateMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyMenu(event *linebot.Event, bot *linebot.Client) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeMenuTextMessage(), makeMenuMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyLocation(event *linebot.Event, bot *linebot.Client) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeReservationTimeMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyConfirmation(event *linebot.Event, bot *linebot.Client) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeConfirmationTextMessage(), makeConfirmationButtonMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func replyThankYou(event *linebot.Event, bot *linebot.Client) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeReservationDateMessage()).Do(); err != nil {
		return err
	}
	return nil
}
