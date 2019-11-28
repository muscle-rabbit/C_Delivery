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
	selectProduct
	decideProduct
	selectLocation
	confirm
	end
)

func (app *app) reply(event *linebot.Event, userID string) *appError {
	session := app.sessionStore.searchSession(userID)
	if session == nil {
		session = app.sessionStore.createSession(userID)
		err := app.createOrder(userID)
		// session.orderID = "ki2XibhAyOFt4dIlYzJfXwwcR2LS_WFxszkIh7-QhV0"

		if err != nil {
			return appErrorf(err, "couldn't create order doc")
		}
	}

	fmt.Println("this is prev step: ", session.prevStep)
	if ok, err := app.validateMessage(event, getMessageFromEvent(event), userID); !ok || (err != nil) {
		return appErrorf(err, "couldn't validate message: %v", err)
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
	case setMenu, selectProduct:
		if err := app.waitProductSelection(event, userID); err != nil {
			return appErrorf(err, "couldn't recieve selected Product")
		}
	case selectLocation:
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

	if err := app.updateOrderInChat(userID, Order{Date: getMessageFromEvent(event)}); err != nil {
		return err
	}
	session.prevStep = reservateTime

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, app.makeReservationTimeMessage()).Do(); err != nil {
		return err
	}
	return nil
}

func (app *app) replyMenu(event *linebot.Event, userID string) error {
	message := getMessageFromEvent(event)
	userSession := app.sessionStore.searchSession(userID)

	// メニューカルセールを返す。
	app.updateOrderInChat(userID, Order{Time: message})
	flexMessage, err := app.makeMenuMessage()
	if err != nil {
		return err
	}

	userSession.prevStep = setMenu
	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeMenuTextMessage(), flexMessage).Do(); err != nil {
		return err
	}
	return nil
}

func (app *app) waitProductSelection(event *linebot.Event, userID string) error {
	message := getMessageFromEvent(event)
	userSession := app.sessionStore.searchSession(userID)

	userSession.prevStep = selectProduct

	if message == "注文決定" {
		if len(userSession.products) == 0 {
			_, err := app.bot.client.ReplyMessage(event.ReplyToken, makeUnselectedProductsMessage()).Do()
			return err
		}
		userSession.prevStep = decideProduct
		if err := app.updateOrderInChat(userID, Order{TotalPrice: app.service.menu.calcPrice(userSession.products)}); err != nil {
			return err
		}

		// 次のステップに移る。
		if err := app.replyLocation(event, userID); err != nil {
			return err
		}
		return nil
	}

	outOfstock, err := app.reserveProducts(userID, message)
	if err != nil {
		return err
	}
	if outOfstock {
		if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeOutOfStockMessage()).Do(); err != nil {
			return err
		}
	}
	if err := app.updateOrderInChat(userID, Order{Products: userSession.products}); err != nil {
		return err
	}

	return err
}

func (app *app) replyLocation(event *linebot.Event, userID string) error {
	session := app.sessionStore.searchSession(userID)

	session.prevStep = selectLocation

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, app.makeLocationMessage()).Do(); err != nil {
		return err
	}
	return nil
}
func (app *app) replyConfirmation(event *linebot.Event, userID string) error {
	userSession := app.sessionStore.searchSession(userID)

	// 一つ前のステップで取得した値をセットする。
	if err := app.updateOrderInChat(userID, Order{Location: getMessageFromEvent(event)}); err != nil {
		return err
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
	message := getMessageFromEvent(event)
	if message == "はい" {
		if err := app.replyThankYou(event, userID); err != nil {
			return err
		}
	} else {
		if err := app.replySorry(event, userID, "注文内容に誤りがあったため"); err != nil {
			return err
		}

	}
	return nil
}

func (app *app) replyThankYou(event *linebot.Event, userID string) error {
	if err := app.completeOrderInChat(userID); err != nil {
		return err
	}

	message, err := app.makeOrderDetail(userID)
	if err != nil {
		return err
	}

	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeThankYouMessage(), message).Do(); err != nil {
		return err
	}
	if err := app.sessionStore.deleteUserSession(userID); err != nil {
		return err
	}

	return nil
}
func (app *app) replySorry(event *linebot.Event, userID string, cause string) error {
	if _, err := app.bot.client.ReplyMessage(event.ReplyToken, makeSorryMessage(cause)).Do(); err != nil {
		return err
	}

	if err := app.cancelOrder(userID); err != nil {
		return err
	}

	return nil
}

func (app *app) replyDenyWorkerLogin(event *linebot.Event, userID string) error {
	message, err := app.makeDenyWorkerMessage(userID)
	if err != nil {
		return err
	}
	_, err = app.bot.client.ReplyMessage(event.ReplyToken, message).Do()
	return err
}

func (app *app) replyWorkerPanel(event *linebot.Event, userID string) error {
	_, err := app.bot.client.ReplyMessage(event.ReplyToken, makeWorkerPanelMessage(userID)).Do()
	return err
}

func (app *app) replyAskCorrectMessages(event *linebot.Event) error {
	_, err := app.bot.client.ReplyMessage(event.ReplyToken, makeAskCorrectMessages()).Do()
	return err
}

func parseMessageToProductText(text string, menu Menu) (Products, error) {
	products := make(Products)
	product, err := splitProductMessage(text, menu)
	if err != nil {
		return nil, err
	}
	productDoc, err := menu.searchProductByName(product.Name)

	products[productDoc.ID] = &Product{Name: product.Name, Stock: product.Stock, Reserved: false}
	return products, nil
}

func splitProductMessage(text string, menu Menu) (*Product, error) {
	i := strings.Index(text, "x")
	n, err := strconv.Atoi(string(text[i+1:]))
	if err != nil {
		return &Product{}, fmt.Errorf("couldn't convert string to int: %v", err)
	}
	return &Product{Name: string(text[:i]), Stock: n}, nil
}

func (app *app) validateMessage(event *linebot.Event, text string, userID string) (bool, error) {
	var ok bool
	var err error
	userSession := app.sessionStore.sessions[userID]
	switch userSession.prevStep {
	case begin:
		ok = validateBegin(text)
		break
	case reservateDate:
		ok, err = validateDate(text)
		break
	case reservateTime:
		ok, err = validateTime(text)
		break
	case setMenu, selectProduct:
		ok, err = validateProduct(text)
		break
	case selectLocation:
		ok = validateLocation(text, app.service.locations)
		break
	default:
		err = fmt.Errorf("couldn't match any prevStep")
	}
	if !ok {
		err := app.replyAskCorrectMessages(event)
		return ok, err
	}
	return ok, err
}

func getMessageFromEvent(event *linebot.Event) string {
	return event.Message.(*linebot.TextMessage).Text
}
