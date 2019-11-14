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

func (app *app) reply(event *linebot.Event, user User) *appError {
	session := app.sessionStore.searchSession(user.UserID)
	if session == nil {
		session = app.sessionStore.createSession(user.UserID)
		// err := app.createOrder(user)
		session.orderID = "4KjVlBdMrY61z_QhPe1q9vpwPNYkJ6rfYxh0XaEkllQ"

		// if err != nil {
		// 	return appErrorf(err, "couldn't create order doc")
		// }
	}

	if !app.sessionStore.checkSessionLifespan(user.UserID) {
		err := app.replySorry(event, user.UserID, "一回の注文にかけられる時間（10分）の上限に達したため")
		if err != nil {
			return appErrorf(err, "couldn't reply sorry")
		}
		return nil

	}

	switch session.prevStep {
	case begin:
		if err := app.replyReservationDate(event, user.UserID); err != nil {
			return appErrorf(err, "couldn't reply ReservationDate: %v", err)
		}
	case reservateDate:
		if err := app.replyReservationTime(event, user.UserID); err != nil {
			return appErrorf(err, "couldn't reply ReservationTime")
		}
	case reservateTime:
		if err := app.replyMenu(event, user.UserID); err != nil {
			return appErrorf(err, "couldn't reply Menu")
		}
	case setMenu:
		if err := app.replyHalfConfirmation(event, user.UserID); err != nil {
			return appErrorf(err, "couldn't reply location")
		}
	case setLocation:
		if err := app.replyConfirmation(event, user.UserID); err != nil {
			return appErrorf(err, "couldn't reply confirmation")
		}
	case confirm, end:
		if err := app.replyFinalMessage(event, user.UserID); err != nil {
			return appErrorf(err, "couldn't reply thankyou")
		}
	default:
		if err := app.replySorry(event, user.UserID, "注文内容に誤りがあった"); err != nil {
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
		if err := app.updateOrderInChat(userID, Order{Date: message.Text}); err != nil {
			return err
		}
	}

	session.prevStep = reservateTime

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, app.makeReservationTimeMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func (app *app) replyMenu(event *linebot.Event, userID string) error {
	userSession := app.sessionStore.searchSession(userID)

	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if isTimeMessage(message.Text) {
			// メニューカルセールを返す。
			app.updateOrderInChat(userID, Order{Time: message.Text})
			message, err := app.makeMenuMessage()
			if err != nil {
				return err
			}
			if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeMenuTextMessage(), message).Do(); err != nil {
				return err
			}
		} else if message.Text == "注文決定" {
			// 次のステップに移る。
			userSession.prevStep = setMenu
			message, err := app.makeHalfConfirmation(userID)
			if err != nil {
				return err
			}

			if _, err := app.bot.client.ReplyMessage(event.ReplyToken, message, makeConfirmationButtonMessage()).Do(); err != nil {
				return err
			}
		} else {
			// 注文メッセージを待ち受ける。 expeted: {商品名} × n
			if err := userSession.products.parseMessageToProductText(message.Text, app.service.menu); err != nil {
				return err
			}
			if err := app.reserveProducts(userID); err != nil {
				return err
			}
			if err := app.updateOrderInChat(userID, Order{Products: userSession.products}); err != nil {
				return err
			}
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

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, app.makeLocationMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func (app *app) replyConfirmation(event *linebot.Event, userID string) error {
	userSession := app.sessionStore.searchSession(userID)

	// 一つ前のステップで取得した値をセットする。
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if err := app.updateOrderInChat(userID, Order{Location: message.Text}); err != nil {
			return err
		}
	}
	userSession.prevStep = confirm

	message, err := app.makeConfirmationTextMessage(userID)
	if err != nil {
		return err
	}

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, message, makeConfirmationButtonMessage()).Do(); err != nil {
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
	if err := app.completeOrderInChat(userID); err != nil {
		return err
	}

	if err := app.sessionStore.deleteUserSession(userID); err != nil {
		return err
	}

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeThankYouMessage()).Do(); err != nil {
		return err
	}

	return nil
}
func (app *app) replySorry(event *linebot.Event, userID string, cause string) error {
	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeSorryMessage(cause), makeReservationDateMessage()).Do(); err != nil {
		return err
	}

	if err := app.cancelOrder(userID); err != nil {
		return err
	}

	return nil
}

func isTimeMessage(text string) bool {
	timeFormat := "~"
	return strings.Contains(text, timeFormat)
}

func (products Products) parseMessageToProductText(text string, menu Menu) error {
	i := strings.Index(text, "x")
	n, err := strconv.Atoi(string(text[i+1:]))
	if err != nil {
		return fmt.Errorf("couldn't convert string to int: %v", err)
	}
	id := menu.searchItemIDByName(string(text[:i]))
	if products[id] != nil {
		product := Product{Name: menu.searchItemNameByID(id), Stock: products[id].Stock + n, Reserved: false}
		products[id] = &product
		return nil
	}
	product := Product{Name: menu.searchItemNameByID(id), Stock: n, Reserved: false}
	products[id] = &product
	return nil
}
