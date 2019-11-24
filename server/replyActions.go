package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/shinyamizuno1008/C_Delivery/server/firestore"
)

func (client *Client) reply(event *linebot.Event, userID string) *appError {
	session := client.sessionStore.searchSession(userID)
	if session == nil {
		session = client.sessionStore.createSession(userID)
		err := client.createOrder(userID)
		// session.orderID = "ki2XibhAyOFt4dIlYzJfXwwcR2LS_WFxszkIh7-QhV0"

		if err != nil {
			return appErrorf(err, "couldn't create order doc")
		}
	}

	switch session.prevStep {
	case begin:
		if err := client.replyReservationDate(event, userID); err != nil {
			return appErrorf(err, "couldn't reply ReservationDate: %v", err)
		}
	case reservateDate:
		if err := client.replyReservationTime(event, userID); err != nil {
			return appErrorf(err, "couldn't reply ReservationTime")
		}
	case reservateTime:
		if err := client.replyMenu(event, userID); err != nil {
			return appErrorf(err, "couldn't reply Menu")
		}
	case setMenu:
		if err := client.replyHalfConfirmation(event, userID); err != nil {
			return appErrorf(err, "couldn't reply location")
		}
	case setLocation:
		if err := client.replyConfirmation(event, userID); err != nil {
			return appErrorf(err, "couldn't reply confirmation")
		}
	case confirm, end:
		if err := client.replyFinalMessage(event, userID); err != nil {
			return appErrorf(err, "couldn't reply thankyou")
		}
	default:
		if err := client.replySorry(event, userID, "注文内容に誤りがあった"); err != nil {
			return appErrorf(err, "couldn't reply sorry")
		}
	}
	return nil
}

func (client *Client) replyReservationDate(event *linebot.Event, userID string) error {
	session := client.sessionStore.searchSession(userID)
	session.prevStep = reservateDate

	if _, err := client.bot.client.ReplyMessage(event.ReplyToken, makeReservationDateMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func (client *Client) replyReservationTime(event *linebot.Event, userID string) error {
	session := client.sessionStore.searchSession(userID)

	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if err := client.updateOrderInChat(userID, Order{Date: message.Text}); err != nil {
			return err
		}
	}

	session.prevStep = reservateTime

	if _, err := client.ReplyMessage(event.ReplyToken, client.makeReservationTimeMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func (client *Client) replyMenu(event *linebot.Event, userID string) error {
	userSession := client.sessionStore.searchSession(userID)

	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if isTimeMessage(message.Text) {
			// メニューカルセールを返す。
			client.updateOrderInChat(userID, Order{Time: message.Text})
			message, err := client.makeMenuMessage()
			if err != nil {
				return err
			}
			if _, err := client.bot.client.ReplyMessage(event.ReplyToken, makeMenuTextMessage(), message).Do(); err != nil {
				return err
			}
		} else if message.Text == "注文決定" {
			if len(userSession.products) == 0 {
				_, err := client.bot.client.ReplyMessage(event.ReplyToken, makeUnselectedProductsMessage()).Do()
				return err
			}
			// 次のステップに移る。
			userSession.prevStep = setMenu
			message, err := client.makeHalfConfirmation(userID)
			if err != nil {
				return err
			}

			if _, err := client.bot.client.ReplyMessage(event.ReplyToken, message, makeConfirmationButtonMessage()).Do(); err != nil {
				return err
			}
		} else {
			// 注文メッセージを待ち受ける。 expeted: {商品名} × n
			outOfstock, err := client.reserveProducts(userID, message.Text)
			if err != nil {
				return err
			}
			if outOfstock {
				if _, err := client.bot.client.ReplyMessage(event.ReplyToken, makeOutOfStockMessage()).Do(); err != nil {
					return err
				}
			}
			if err := client.updateOrderInChat(userID, Order{Products: userSession.products}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (client *Client) replyHalfConfirmation(event *linebot.Event, userID string) error {
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。

	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if message.Text == "はい" {
			if err := client.replyLocation(event, userID); err != nil {
				return err
			}
		} else {
			if err := client.replySorry(event, userID, "注文内容に誤りがあったため"); err != nil {
				return err
			}

		}
	}
	return nil
}

func (client *Client) replyLocation(event *linebot.Event, userID string) error {
	session := client.sessionStore.searchSession(userID)

	session.prevStep = setLocation

	if _, err := client.bot.client.ReplyMessage(event.ReplyToken, client.makeLocationMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func (client *Client) replyConfirmation(event *linebot.Event, userID string) error {
	userSession := client.sessionStore.searchSession(userID)

	// 一つ前のステップで取得した値をセットする。
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if err := client.updateOrderInChat(userID, Order{Location: message.Text}); err != nil {
			return err
		}
	}
	userSession.prevStep = confirm

	message, err := client.makeConfirmationTextMessage(userID)
	if err != nil {
		return err
	}

	if _, err := client.bot.client.ReplyMessage(event.ReplyToken, message, makeConfirmationButtonMessage()).Do(); err != nil {
		return err
	}
	return nil
}

func (client *Client) replyFinalMessage(event *linebot.Event, userID string) error {
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if message.Text == "はい" {
			if err := client.replyThankYou(event, userID); err != nil {
				return err
			}
		} else {
			if err := client.replySorry(event, userID, "注文内容に誤りがあったため"); err != nil {
				return err
			}

		}
	}
	return nil
}

func (client *Client) replyThankYou(event *linebot.Event, userID string) error {
	if err := client.completeOrderInChat(userID); err != nil {
		return err
	}

	message, err := client.makeOrderDetail(userID)
	if err != nil {
		return err
	}

	if _, err := client.Bot.ReplyMessage(event.ReplyToken, makeThankYouMessage(), message).Do(); err != nil {
		return err
	}
	if err := client.sessionStore.deleteUserSession(userID); err != nil {
		return err
	}

	return nil
}
func (client *Client) replySorry(event *linebot.Event, userID string, cause string) error {
	if _, err := client.Bot.ReplyMessage(event.ReplyToken, makeSorryMessage(cause)).Do(); err != nil {
		return err
	}

	if err := client.cancelOrder(userID); err != nil {
		return err
	}

	return nil
}

func (client *Client) replyDenyWorkerLogin(event *linebot.Event, userID string) error {
	message, err := client.makeDenyWorkerMessage(userID)
	if err != nil {
		return err
	}
	_, err = client.Bot.ReplyMessage(event.ReplyToken, message).Do()
	return err
}

func (client *Client) replyWorkerPanel(event *linebot.Event, userID string) error {
	_, err := client.Bot.ReplyMessage(event.ReplyToken, makeWorkerPanelMessage(userID)).Do()
	return err
}

func isTimeMessage(text string) bool {
	timeFormat := "~"
	return strings.Contains(text, timeFormat)
}

func parseMessageToProductText(text string, menu firestore.Menu) (firestore.Products, error) {
	p := make(Products)
	i := strings.Index(text, "x")
	n, err := strconv.Atoi(string(text[i+1:]))
	if err != nil {
		return nil, fmt.Errorf("couldn't convert string to int: %v", err)
	}
	id := menu.searchItemIDByName(string(text[:i]))

	p[id] = &Product{Name: menu.searchItemNameByID(id), Stock: n, Reserved: false}
	return p, nil
}
