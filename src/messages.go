package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
)

type Item struct {
	Name     string
	Price    int
	ImageURL string
}

// makeReservationDateMessage は 注文日程指定用のメッセージを返すメソッドです。
func makeReservationDateMessage() *linebot.TemplateMessage {
	// 現状対応できるのは注文日を含めた二日間のみ。従業員が増えたら随時更新していく。 (2019/10/28)
	wdays := [...]string{"日", "月", "火", "水", "木", "金", "土"}
	title := "日程指定"
	phrase := "ご注文日をお選びください。"
	now := time.Now()
	today := fmt.Sprintf("本日 %d/%d (%s)", now.Month(), now.Day(), wdays[now.Weekday()])
	tomorrow := fmt.Sprintf("明日 %d/%d (%s)", now.Month(), now.Day()+1, wdays[now.Weekday()+1])

	template := linebot.NewButtonsTemplate(
		"", title, phrase,
		linebot.NewMessageAction(today, today),
		linebot.NewMessageAction(tomorrow, tomorrow),
	)

	message := linebot.NewTemplateMessage("日程指定", template)
	return message
}

// makeReservationTimeMessage は 注文時間指定用のメッセージを返すメソッドです。
func makeReservationTimeMessage() *linebot.TemplateMessage {
	lastOrder := "12:30"
	title := "時間指定"
	phrase := "ご注文時間をお選びください。\nラストオーダー: " + lastOrder

	template := linebot.NewButtonsTemplate(
		"", title, phrase,
		linebot.NewMessageAction("12:00~12:30", "12:00~12:30"),
		linebot.NewMessageAction("12:30~13:00", "12:30~13:00"),
		linebot.NewMessageAction("13:30~14:00", "13:30~14:00"),
		linebot.NewMessageAction("14:30~15:00", "14:30~15:00"),
	)

	message := linebot.NewTemplateMessage("日程指定", template)
	return message
}

// makeMenuText は 商品指定用のテキストメッセージを返すメソッドです。
func makeMenuTextMessage() *linebot.TextMessage {
	return linebot.NewTextMessage("ご注文品をお選びください。")
}

// makeMenu は 商品指定用の写真付きカルーセルを返すメソッドです。
func makeMenuMessage() *linebot.TemplateMessage {
	items := []Item{
		{"鳥唐揚弁当", 360, "https://takuma-life.jp/wp-content/uploads/2018/05/IMG_1506-1.jpg"},
		{"のり弁当", 300, "https://cdn-ak.f.st-hatena.com/images/fotolife/p/pegaman/20190119/20190119204845.jpg"},
		{"シャケ弁当", 400, "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRMaV54QCJVdR1ZzHIcw2EMvehZEf_5KiizJhY7B_BvqDlGSklI&s"},
		{"烏龍茶", 150, "https://i.ibb.co/QNzQBRn/Screen-Shot-2019-10-28-at-12-43-51.png"},
		{"コカコーラ", 150, "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcScSLdxny37CL0tK4ADpnEVPjwX5jVWuHpxVmfcCt1DSreBG7iF1A&s"},
	}
	marginLeftForName := "                "
	marginLeftForPrice := "                       "
	var columns []*linebot.CarouselColumn
	for _, item := range items {
		columns = append(columns, linebot.NewCarouselColumn(
			item.ImageURL, fmt.Sprintf("%s%s", marginLeftForName, item.Name), fmt.Sprintf("%s¥%s", marginLeftForPrice, strconv.Itoa(item.Price)),
			linebot.NewMessageAction("これにする", item.Name),
		))

	}
	template := linebot.NewCarouselTemplate(
		columns...,
	)
	return linebot.NewTemplateMessage("メニュー指定", template)
}

// makeLocation は 発送先用のメッセージを返すメソッドです。
func makeLocationMessage() *linebot.TemplateMessage {
	locations := []string{"8号間 2F 中央広場"}
	title := "発送先指定"
	phrase := "発送先を選択下さい。"

	// このエラーが解けない
	// cannot use actions (variable of type []*linebot.MessageAction) as []linebot.TemplateAction value in argument to linebot.NewButtonsTemplate

	// var actions []*linebot.MessageAction
	// for _, location := range locations {
	// 	actions = append(actions,
	// 		linebot.NewMessageAction(location, location),
	// 	)
	// }

	// template := linebot.NewButtonsTemplate(
	// 	"", title, phrase,
	// 	actions...,
	// )

	template := linebot.NewButtonsTemplate(
		"", title, phrase,
		linebot.NewMessageAction(locations[0], locations[0]),
	)
	return linebot.NewTemplateMessage(title, template)
}

type Order struct {
	date     string
	time     string
	location string
	items    []Item
}

// makeConfirmationText は 注文確認テキスト用メッセージを送信するメソッドです。
func makeConfirmationTextMessage() *linebot.TextMessage {
	// TODO: あとで消して、注文データはデータベースに保存するようにする。
	order := Order{"11/1", "12:00~12:30", "8号館中央広場", []Item{
		{"鳥唐揚弁当", 360, "https://takuma-life.jp/wp-content/uploads/2018/05/IMG_1506-1.jpg"},
		{"のり弁当", 300, "https://cdn-ak.f.st-hatena.com/images/fotolife/p/pegaman/20190119/20190119204845.jpg"},
		{"シャケ弁当", 400, "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRMaV54QCJVdR1ZzHIcw2EMvehZEf_5KiizJhY7B_BvqDlGSklI&s"},
		{"烏龍茶", 150, "https://i.ibb.co/QNzQBRn/Screen-Shot-2019-10-28-at-12-43-51.png"},
		{"コカコーラ", 150, "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcScSLdxny37CL0tK4ADpnEVPjwX5jVWuHpxVmfcCt1DSreBG7iF1A&s"},
	}}
	// TODO: あとで消す
	var menu string
	var price int

	for i, item := range order.items {
		if i == 0 {
			menu = item.Name
		} else {
			menu += ", " + item.Name
		}
		price += item.Price
	}
	orderDetail := fmt.Sprintf("ご注文内容確認\n\n1. 日時 : %s\n2. 時間 : %s\n3. 発送場所: %s\n3. 品物 : %s\n4. お会計: ¥%d",
		order.date, order.time, order.location, menu, price)

	return linebot.NewTextMessage(orderDetail)
}

// makeConfirmationButton は 注文確認テキスト用ボタンを送信するメソッドです。
func makeConfirmationButtonMessage() *linebot.TemplateMessage {

	title := "ご注文は、こちらでお間違いありませんか？"
	confirmationTemplate := linebot.NewConfirmTemplate(
		title,
		linebot.NewMessageAction("はい", "はい"),
		linebot.NewMessageAction("いいえ", "いいえ"),
	)

	return linebot.NewTemplateMessage("ご注文確認", confirmationTemplate)
}

// makeThankYou は お礼メッセージを送信するメソッドです。
func makeThankYouMessage() *linebot.TextMessage {
	message := "ご注文ありがとうございました。\n\n当日は現金をご用意の上\n所定の場所にお集まりください。\n\nまたのご利用お待ちしております。"
	return linebot.NewTextMessage(message)
}
