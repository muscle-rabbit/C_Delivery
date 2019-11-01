package main

import (
	"encoding/json"
	"fmt"
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

func (client client) reply(event *linebot.Event) *appError {
	prevStep := client.session.Values["prev_step"]
	switch prevStep {
	case nil, begin:
		if err := replyReservationDate(event, client.bot, client); err != nil {
			return appErrorf(err, "couldn't reply ReservationDate")
		}
	case reservationData:
		if err := replyReservationTime(event, client.bot, client); err != nil {
			return appErrorf(err, "couldn't reply ReservationTime")
		}
	case reservationTime:
		if err := replyMenu(event, client.bot, client); err != nil {
			return appErrorf(err, "couldn't reply Menu")
		}
	case menu:
		if err := replyHalfConfirmation(event, client.bot, client); err != nil {
			return appErrorf(err, "couldn't reply location")
		}
	case location:
		if err := replyConfirmation(event, client.bot, client); err != nil {
			return appErrorf(err, "couldn't reply confirmation")
		}
	case confirmation, end:
		if err := replyFinalMessage(event, client.bot, client); err != nil {
			return appErrorf(err, "couldn't reply thankyou")
		}
	default:
		if err := replySorry(event, client.bot, client); err != nil {
			return appErrorf(err, "couldn't reply sorry")
		}
	}
	return nil
}

func replyReservationDate(event *linebot.Event, bot *linebot.Client, client client) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeReservationDateMessage()).Do(); err != nil {
		return err
	}

	client.session.Values["prev_step"] = reservationData
	if err := client.session.Save(client.request, client.writer); err != nil {
		return err
	}
	return nil
}
func replyReservationTime(event *linebot.Event, bot *linebot.Client, client client) error {
	ot := orderTime{detailTime{12, 00}, detailTime{15, 00}, 30, "12:30"}
	if _, err := bot.ReplyMessage(event.ReplyToken, makeReservationTimeMessage(ot.makeTimeTable(), ot.lastorder)).Do(); err != nil {
		return err
	}

	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		order := Order{
			Date: message.Text,
		}
		jorder, _ := json.Marshal(order)
		client.session.Values["order"] = jorder
	}

	client.session.Values["prev_step"] = reservationTime
	if err := client.session.Save(client.request, client.writer); err != nil {
		return err
	}
	return nil
}
func replyMenu(event *linebot.Event, bot *linebot.Client, client client) error {

	// session から ユーザーの選択中の情報を取得。
	var userOrder Order
	if b, ok := client.session.Values["order"].([]byte); ok {
		json.Unmarshal(b, &userOrder)
	} else {
		return fmt.Errorf("couldn't do type assertion.")
	}

	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if isTimeMessage(message.Text) {
			userOrder.Time = message.Text
			jorder, _ := json.Marshal(userOrder)
			client.session.Values["order"] = jorder
			if _, err := bot.ReplyMessage(event.ReplyToken, makeMenuTextMessage(), makeMenuMessage()).Do(); err != nil {
				return err
			}
		} else if message.Text == "注文決定" {
			client.session.Values["prev_step"] = menu
			jorder, _ := json.Marshal(userOrder)
			client.session.Values["order"] = jorder
			if _, err := bot.ReplyMessage(event.ReplyToken, makeHalfConfirmation(userOrder), makeConfirmationButtonMessage()).Do(); err != nil {
				return err
			}
		} else {
			userOrder.MenuList = append(userOrder.MenuList, menuList.searchItemByName(message.Text))
			jorder, _ := json.Marshal(userOrder)
			client.session.Values["order"] = jorder
		}
	}

	if err := client.session.Save(client.request, client.writer); err != nil {
		return err
	}

	return nil
}

func replyHalfConfirmation(event *linebot.Event, bot *linebot.Client, client client) error {
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if message.Text == "はい" {
			if err := replyLocation(event, bot, client); err != nil {
				return err
			}
		} else {
			if err := replySorry(event, bot, client); err != nil {
				return err
			}

		}
	}
	return nil

}

func replyLocation(event *linebot.Event, bot *linebot.Client, client client) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeLocationMessage()).Do(); err != nil {
		return err
	}
	client.session.Values["prev_step"] = location
	if err := client.session.Save(client.request, client.writer); err != nil {
		return err
	}
	return nil
}
func replyConfirmation(event *linebot.Event, bot *linebot.Client, client client) error {

	// session から ユーザーの選択中の情報を取得。
	var userOrder Order
	if b, ok := client.session.Values["order"].([]byte); ok {
		json.Unmarshal(b, &userOrder)
	} else {
		return fmt.Errorf("couldn't do type assertion.")
	}

	// 一つ前のステップで取得した値をセットする。
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		userOrder.Location = message.Text
		jorder, err := json.Marshal(userOrder)
		if err != nil {
			return fmt.Errorf("couldn't make userOrder json: %v", jorder)
		}
		client.session.Values["order"] = jorder
	}

	if _, err := bot.ReplyMessage(event.ReplyToken, makeConfirmationTextMessage(userOrder), makeConfirmationButtonMessage()).Do(); err != nil {
		return err
	}
	client.session.Values["prev_step"] = confirmation

	if err := client.session.Save(client.request, client.writer); err != nil {
		return err
	}
	return nil
}

func replyFinalMessage(event *linebot.Event, bot *linebot.Client, client client) error {
	// TODO: 冗長なのでリファクタ必要。event.Message.Text みたいな使い方したい。
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		if message.Text == "はい" {
			if err := replyThankYou(event, bot, client); err != nil {
				return err
			}
		} else {
			if err := replySorry(event, bot, client); err != nil {
				return err
			}

		}
	}
	return nil
}

func replyThankYou(event *linebot.Event, bot *linebot.Client, client client) error {
	if _, err := bot.ReplyMessage(event.ReplyToken, makeThankYouMessage()).Do(); err != nil {
		return err
	}

	// session から ユーザーの選択中の情報を取得。
	var userOrder Order
	if b, ok := client.session.Values["order"].([]byte); ok {
		json.Unmarshal(b, &userOrder)
	} else {
		return fmt.Errorf("couldn't do type assertion.")
	}

	// client.session.Values["prev_step"] = end
	client.session.Values["prev_step"] = begin
	if err := client.session.Save(client.request, client.writer); err != nil {
		return err
	}
	return nil
}
func replySorry(event *linebot.Event, bot *linebot.Client, client client) error {
	// TODO: session を破棄
	if _, err := bot.ReplyMessage(event.ReplyToken, makeSorryMessage(), makeReservationDateMessage()).Do(); err != nil {
		return err
	}

	client.session.Values["prev_step"] = begin
	if err := client.session.Save(client.request, client.writer); err != nil {
		return err
	}
	return nil
}

func isTimeMessage(text string) bool {
	timeFormat := "~"
	return strings.Contains(text, timeFormat)
}
