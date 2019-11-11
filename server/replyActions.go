package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	begin int = iota
	reservateDate
	reservateTime
	setMenu
	setLocation
	confirm
	end
)

func (app *app) reply(event *linebot.Event, userID string) *appError {
	session := app.sessionStore.searchSession(userID)
	if session == nil {
		session = app.sessionStore.createSession(userID)
		// err := app.createOrder(userID)
		session.orderID = "VvPwHOtxqO99QrVPPgYXiBWVKbyYD7e85p4B68QmZqY"

		// if err != nil {
		// 	return appErrorf(err, "couldn't create order doc")
		// }
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
	case reservateDate:
		if err := app.replyReservationTime(event, userID); err != nil {
			return appErrorf(err, "couldn't reply ReservationTime")
		}
	case reservateTime:
		if err := app.replyMenu(event, userID); err != nil {
			return appErrorf(err, "couldn't reply Menu")
		}
	case setMenu:
		if err := app.replyHalfConfirmation(event, userID); err != nil {
			return appErrorf(err, "couldn't reply location")
		}
	case setLocation:
		if err := app.replyConfirmation(event, userID); err != nil {
			return appErrorf(err, "couldn't reply confirmation")
		}
	case confirm, end:
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
	session.prevStep = reservateDate

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeReservationDateMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func (app *app) replyReservationTime(event *linebot.Event, userID string) error {
	session := app.sessionStore.searchSession(userID)

	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if err := app.updateOrderInChat(userID, Order{
			Date: message.Text,
		}, session.prevStep); err != nil {
			return err
		}
	}

	session.prevStep = reservateTime
	bh := app.service.businessHours

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeReservationTimeMessage(bh.makeTimeTable(), bh.lastorder)).Do(); err != nil {
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
			// メニューカルセールを返す。
			app.updateOrderInChat(userID, Order{Time: message.Text}, session.prevStep)
			if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeMenuTextMessage(), makeMenuMessage(app.service.menu)).Do(); err != nil {
				return err
			}
		} else if message.Text == "注文決定" {
			// 次のステップに移る。
			session.prevStep = setMenu
			price := app.service.menu.calcPrice(session.products)

			order, err := app.fetchUserOrder(userID)
			if err != nil {
				return err
			}

			if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeHalfConfirmation(session.products, app.service.menu, order, price), makeConfirmationButtonMessage()).Do(); err != nil {
				return err
			}
		} else {
			// 注文メッセージを待ち受ける。 expeted: {商品名} × n
			if err := session.products.parseProductsText(message.Text, app.service.menu); err != nil {
				return err
			}
			app.updateOrderInChat(userID, Order{Products: session.products}, session.prevStep)
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

	session.prevStep = setLocation

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeLocationMessage(app.service.locations)).Do(); err != nil {
		return err
	}
	return nil
}
func (app *app) replyConfirmation(event *linebot.Event, userID string) error {
	session := app.sessionStore.searchSession(userID)

	// 一つ前のステップで取得した値をセットする。
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if err := app.updateOrderInChat(userID, Order{Location: message.Text}, session.prevStep); err != nil {
			return err
		}
	}
	session.prevStep = confirm

	price := app.service.menu.calcPrice(session.products)
	order, err := app.fetchUserOrder(userID)
	if err != nil {
		return err
	}
	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeConfirmationTextMessage(session.products, app.service.menu, order, price), makeConfirmationButtonMessage()).Do(); err != nil {
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

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeThankYouMessage()).Do(); err != nil {
		return err
	}

	return nil
}
func (app *app) replySorry(event *linebot.Event, userID string, cause string) error {
	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeSorryMessage(cause), makeReservationDateMessage()).Do(); err != nil {
		return err
	}

	if err := app.deleteOrder(userID); err != nil {
		return err
	}
	app.sessionStore.deleteUserSession(userID)
	return nil
}

func isTimeMessage(text string) bool {
	timeFormat := "~"
	return strings.Contains(text, timeFormat)
}

func (products Products) parseProductsText(text string, menu Menu) error {
	i := strings.Index(text, "x")
	d, err := strconv.Atoi(string(text[i+1:]))
	if err != nil {
		return fmt.Errorf("couldn't convert string to int: %v", err)
	}

	products[menu.searchItemIDByName(string(text[:i]))] = d
	return nil
}
