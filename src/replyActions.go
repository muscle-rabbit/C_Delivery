package main

import (
	"strings"

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

func (app *app) reply(event *linebot.Event, userID string) *appError {
	session := app.sessionStore.searchSession(userID)
	if session == nil {
		session = app.sessionStore.createSession(userID)
	}

	if !app.sessionStore.checkSessionLifespan(userID) {
		err := app.replySorry(event, userID, "一回の注文にかけられる時間（10分）の上限に達したため")
		if err != nil {
			return appErrorf(err, "couldn't reply sorry")
		}
		return nil

	}

	switch session.prevStep {
	case begin:
		if err := app.replyReservationDate(event, userID); err != nil {
			return appErrorf(err, "couldn't reply ReservationDate: %v", err)
		}
	case reservationData:
		if err := app.replyReservationTime(event, userID); err != nil {
			return appErrorf(err, "couldn't reply ReservationTime")
		}
	case reservationTime:
		if err := app.replyMenu(event, userID); err != nil {
			return appErrorf(err, "couldn't reply Menu")
		}
	case menu:
		if err := app.replyHalfConfirmation(event, userID); err != nil {
			return appErrorf(err, "couldn't reply location")
		}
	case location:
		if err := app.replyConfirmation(event, userID); err != nil {
			return appErrorf(err, "couldn't reply confirmation")
		}
	case confirmation, end:
		if err := app.replyFinalMessage(event, userID); err != nil {
			return appErrorf(err, "couldn't reply thankyou")
		}
	default:
		if err := app.replySorry(event, userID, "注文内容に誤りがあった"); err != nil {
			return appErrorf(err, "couldn't reply sorry")
		}
	}
	return nil
}

func (app *app) replyReservationDate(event *linebot.Event, userID string) error {
	session := app.sessionStore.searchSession(userID)
	session.prevStep = reservationData

	if _, err := app.bot.ReplyMessage(event.ReplyToken, makeReservationDateMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func (app *app) replyReservationTime(event *linebot.Event, userID string) error {
	session := app.sessionStore.searchSession(userID)

	session.prevStep = reservationTime
	ot := orderTime{detailTime{12, 00}, detailTime{15, 00}, 30, "12:30"}

	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		session.order = Order{Date: message.Text}
	}

	if _, err := app.bot.ReplyMessage(event.ReplyToken, makeReservationTimeMessage(ot.makeTimeTable(), ot.lastorder)).Do(); err != nil {
		return err
	}
	return nil
}
func (app *app) replyMenu(event *linebot.Event, userID string) error {
	session := app.sessionStore.searchSession(userID)

	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if isTimeMessage(message.Text) {
			session.order.Time = message.Text
			if _, err := app.bot.ReplyMessage(event.ReplyToken, makeMenuTextMessage(), makeMenuMessage(app.menu)).Do(); err != nil {
				return err
			}
		} else if message.Text == "注文決定" {
			session.prevStep = menu
			if _, err := app.bot.ReplyMessage(event.ReplyToken, makeHalfConfirmation(session.order), makeConfirmationButtonMessage()).Do(); err != nil {
				return err
			}
		} else {
			session.order.MenuList = append(session.order.MenuList, app.menu.searchItemByName(message.Text))
		}
	}
	return nil
}

func (app *app) replyHalfConfirmation(event *linebot.Event, userID string) error {
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。

	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if message.Text == "はい" {
			if err := app.replyLocation(event, userID); err != nil {
				return err
			}
		} else {
			if err := app.replySorry(event, userID, "注文内容に誤りがあったため"); err != nil {
				return err
			}

		}
	}
	return nil

}

func (app *app) replyLocation(event *linebot.Event, userID string) error {
	session := app.sessionStore.searchSession(userID)

	session.prevStep = location

	if _, err := app.bot.ReplyMessage(event.ReplyToken, makeLocationMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func (app *app) replyConfirmation(event *linebot.Event, userID string) error {
	session := app.sessionStore.searchSession(userID)
	session.prevStep = confirmation

	// 一つ前のステップで取得した値をセットする。
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		session.order.Location = message.Text
	}

	if _, err := app.bot.ReplyMessage(event.ReplyToken, makeConfirmationTextMessage(session.order), makeConfirmationButtonMessage()).Do(); err != nil {
		return err
	}
	return nil
}

func (app *app) replyFinalMessage(event *linebot.Event, userID string) error {
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if message.Text == "はい" {
			if err := app.replyThankYou(event, userID); err != nil {
				return err
			}
		} else {
			if err := app.replySorry(event, userID, "注文内容に誤りがあったため"); err != nil {
				return err
			}

		}
	}
	return nil
}

func (app *app) replyThankYou(event *linebot.Event, userID string) error {
	session := app.sessionStore.searchSession(userID)

	session.prevStep = begin
	app.sessionStore.deleteUserSession(userID)

	if _, err := app.bot.ReplyMessage(event.ReplyToken, makeThankYouMessage()).Do(); err != nil {
		return err
	}

	return nil
}
func (app *app) replySorry(event *linebot.Event, userID string, cause string) error {
	if _, err := app.bot.ReplyMessage(event.ReplyToken, makeSorryMessage(cause), makeReservationDateMessage()).Do(); err != nil {
		return err
	}

	app.sessionStore.deleteUserSession(userID)
	return nil
}

func isTimeMessage(text string) bool {
	timeFormat := "~"
	return strings.Contains(text, timeFormat)
}
