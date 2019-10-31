package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
)

type item struct {
	Name     string
	Price    int
	ImageURL string
}

var items = []item{
	{"鳥唐揚弁当", 360, "https://takuma-life.jp/wp-content/uploads/2018/05/IMG_1506-1.jpg"},
	{"のり弁当", 300, "https://cdn-ak.f.st-hatena.com/images/fotolife/p/pegaman/20190119/20190119204845.jpg"},
	{"シャケ弁当", 400, "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRMaV54QCJVdR1ZzHIcw2EMvehZEf_5KiizJhY7B_BvqDlGSklI&s"},
	{"烏龍茶", 150, "https://i.ibb.co/QNzQBRn/Screen-Shot-2019-10-28-at-12-43-51.png"},
	{"コカコーラ", 150, "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcScSLdxny37CL0tK4ADpnEVPjwX5jVWuHpxVmfcCt1DSreBG7iF1A&s"},
}

var wdays = [...]string{"日", "月", "火", "水", "木", "金", "土"}

type oreder struct {
	date     string
	time     string
	location string
	items    []item
}

type orderTime struct {
	begin    detailTime
	end      detailTime
	interval int // minute
}

type detailTime struct {
	hour   int
	minute int
}

func (ot orderTime) makeTimeTable() []string {
	today := time.Now()
	jst, _ := time.LoadLocation("Asia/Tokyo")

	begin := time.Date(today.Year(), today.Month(), today.Day(), ot.begin.hour, ot.begin.minute, 0, 0, jst)
	end := time.Date(today.Year(), today.Month(), today.Day(), ot.end.hour, ot.end.minute, 0, 0, jst)

	diff := end.Sub(begin)
	interval := time.Duration(ot.interval) * time.Minute

	n := int(diff / interval)
	tt := make([]string, n+1)
	layout := "3:04AM"

	for i := 0; i < n+1; i++ {
		if i == 0 {
			tt[0] = begin.Format(layout) + "~" + begin.Add(interval).Format(layout)
			fmt.Println(begin.Format(layout))
			continue
		}
		tt[i] = begin.Add(interval*time.Duration(i-1)).Format(layout) + "~" + begin.Add(interval*time.Duration(i)).Format(layout)
	}

	return tt
}

// makeReservationDateMessage は 注文日程指定用のメッセージを返すメソッドです。
func makeReservationDateMessage() *linebot.TemplateMessage {
	// 現状対応できるのは注文日を含めた二日間のみ。従業員が増えたら随時更新していく。 (2019/10/28)
	title := "日程指定"
	phrase := "ご注文日をお選びください。"

	today := time.Now()

	todayf := fmt.Sprintf("本日 %d/%d (%s)", today.Month(), today.Day(), wdays[today.Weekday()])
	tomorrowf := fmt.Sprintf("明日 %d/%d (%s)", today.Month(), today.Day()+1, wdays[today.Weekday()+1])

	template := linebot.NewButtonsTemplate(
		"", title, phrase,
		linebot.NewMessageAction(todayf, todayf),
		linebot.NewMessageAction(tomorrowf, tomorrowf),
	)

	message := linebot.NewTemplateMessage("日程指定", template)
	return message
}

// makeReservationTimeMessage は 注文時間指定用のメッセージを返すメソッドです。
func makeReservationTimeMessage(timeTable []string) *linebot.TemplateMessage {
	lastoreder := "12:30"
	title := "時間指定"
	phrase := "ご注文時間をお選びください。\nラストオーダー: " + lastoreder

	template := linebot.NewButtonsTemplate(
		"", title, phrase,
		// 4 つ以上の item は追加できない。
		linebot.NewMessageAction(timeTable[0], timeTable[0]),
		linebot.NewMessageAction(timeTable[1], timeTable[1]),
		linebot.NewMessageAction(timeTable[2], timeTable[2]),
		linebot.NewMessageAction(timeTable[3], timeTable[3]),
	)

	message := linebot.NewTemplateMessage("日程指定", template)
	return message
}

// linebot.NewMessageAction("", title, phrase, ...actions) みたいに使いけどできない。
func makeTimeTableMessageAction(timeTable []string) []*linebot.MessageAction {
	actions := make([]*linebot.MessageAction, len(timeTable))
	for i, time := range timeTable {
		actions[i] = linebot.NewMessageAction(time, time)
	}
	return actions
}

// makeMenuText は 商品指定用のテキストメッセージを返すメソッドです。
func makeMenuTextMessage() *linebot.TextMessage {
	return linebot.NewTextMessage("ご注文品をお選びください。")
}

// makeMenu は 商品指定用の写真付きカルーセルを返すメソッドです。
func makeMenuMessage() *linebot.TemplateMessage {
	items := []item{
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

// makeConfirmationText は 注文確認テキスト用メッセージを返すメソッドです。
func makeConfirmationTextMessage() *linebot.TextMessage {
	// TODO: あとで消して、注文データはデータベースに保存するようにする。
	order := oreder{"11/1", "12:00~12:30", "8号館中央広場", []item{
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

// makeConfirmationButton は 注文確認テキスト用ボタンを返すメソッドです。
func makeConfirmationButtonMessage() *linebot.TemplateMessage {

	title := "ご注文は、こちらでお間違いありませんか？"
	confirmationTemplate := linebot.NewConfirmTemplate(
		title,
		linebot.NewMessageAction("はい", "はい"),
		linebot.NewMessageAction("いいえ", "いいえ"),
	)

	return linebot.NewTemplateMessage("ご注文確認", confirmationTemplate)
}

// makeThankYou は お礼メッセージを返するメソッドです。
func makeThankYouMessage() *linebot.TextMessage {
	message := "ご注文ありがとうございました。\n\n当日は現金をご用意の上\n所定の場所にお集まりください。\n\nまたのご利用お待ちしております。"
	return linebot.NewTextMessage(message)
}

// makeSorryMessage は 謝りのメッセージを返すメソッドです。
func makeSorryMessage() *linebot.TextMessage {
	message := "申し訳ありません。\n最初から注文をやり直してください。"
	return linebot.NewTextMessage(message)
}
