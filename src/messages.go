package main

import (
	"bytes"
	"encoding/gob"
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

var menuList = Menu{
	{"鳥唐揚弁当", 360, "https://takuma-life.jp/wp-content/uploads/2018/05/IMG_1506-1.jpg"},
	{"のり弁当", 300, "https://cdn-ak.f.st-hatena.com/images/fotolife/p/pegaman/20190119/20190119204845.jpg"},
	{"シャケ弁当", 400, "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRMaV54QCJVdR1ZzHIcw2EMvehZEf_5KiizJhY7B_BvqDlGSklI&s"},
	{"烏龍茶", 150, "https://i.ibb.co/QNzQBRn/Screen-Shot-2019-10-28-at-12-43-51.png"},
	{"コカコーラ", 150, "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcScSLdxny37CL0tK4ADpnEVPjwX5jVWuHpxVmfcCt1DSreBG7iF1A&s"},
}

type Menu []item

func (m Menu) searchItemByName(key string) item {
	for _, item := range m {
		if item.Name == key {
			return item
		}
	}
	// TODO: nil を返したい。
	return item{}
}

var wdays = [...]string{"日", "月", "火", "水", "木", "金", "土"}
var locations = []string{"8号間 2F 中央広場", "6号館 1F"}

type Order struct {
	Date     string `json:"date"`
	Time     string `json:"time"`
	Location string `json:"location"`
	MenuList Menu   `json:"menuList"`
}

func encodeOrder(order Order) []byte {
	buf := bytes.NewBuffer(nil)
	_ = gob.NewEncoder(buf).Encode(&order)
	return buf.Bytes()
}

func decodeOrder(data []byte) *Order {
	fmt.Println("data in decode", data)
	var o Order
	buf := bytes.NewBuffer(data)
	_ = gob.NewDecoder(buf).Decode(&o)
	return &o

	// return data.(*Order)

	// var buf bytes.Buffer
	// dec := gob.NewDecoder(&buf)
	// dec.Decode(&data)
	// return &Order{}
}

type orderTime struct {
	begin     detailTime
	end       detailTime
	interval  int // minute
	lastorder string
}

type detailTime struct {
	hour   int
	minute int
}

// gorilla/sessions の Values() の返り値用に使う Map 構造体。・
type M map[string]interface{}

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
	// TODO: 現状対応できるのは注文日を含めた二日間のみ。従業員が増えたら随時更新していく。 (2019/10/28)
	title := "日程指定"
	phrase := "ご注文日をお選びください。"

	today := time.Now()

	todayf := fmt.Sprintf("本日 %d/%d (%s)", today.Month(), today.Day(), wdays[today.Weekday()])
	tomorrowf := fmt.Sprintf("明日 %d/%d (%s)", today.Month(), today.Day()+1, wdays[(6+today.Weekday())%6])

	template := linebot.NewButtonsTemplate(
		"", title, phrase,
		linebot.NewMessageAction(todayf, todayf),
		linebot.NewMessageAction(tomorrowf, tomorrowf),
	)

	message := linebot.NewTemplateMessage("日程指定", template)
	return message
}

// makeReservationTimeMessage は 注文時間指定用のメッセージを返すメソッドです。
func makeReservationTimeMessage(timeTable []string, lastorder string) *linebot.TemplateMessage {
	title := "時間指定"
	phrase := "ご注文時間をお選びください。\nラストオーダー: " + lastorder

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
	marginLeftForName := "                "
	marginLeftForPrice := "                       "
	var columns []*linebot.CarouselColumn
	columns = append(columns, linebot.NewCarouselColumn("https://mitkp.com/wp-content/uploads/2017/04/pop_kettei.png", "注文が決まりましたら押してください。", "注文決定",
		linebot.NewMessageAction("これにする", "注文決定")))

	for _, item := range menuList {
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

func makeHalfConfirmation(order Order) *linebot.TextMessage {
	var menuText string
	var priceText int

	for i, item := range order.MenuList {
		if i == 0 {
			menuText = item.Name
		} else {
			menuText += ", " + item.Name
		}
		priceText += item.Price
	}
	return linebot.NewTextMessage(fmt.Sprintf("ご注文途中確認\n\nお間違いなければ次のステップに移ります。\n\n1. 日程 : %s\n2. 時間 : %s\n3. 品物: %s\n\n4. お会計 : %d", order.Date, order.Time, menuText, priceText))
}

// makeLocation は 発送先用のメッセージを返すメソッドです。
func makeLocationMessage() *linebot.TemplateMessage {
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
		linebot.NewMessageAction(locations[1], locations[1]),
	)
	return linebot.NewTemplateMessage(title, template)
}

// makeConfirmationText は 注文確認テキスト用メッセージを返すメソッドです。
func makeConfirmationTextMessage(order Order) *linebot.TextMessage {
	var menuText string
	var priceText int

	for i, item := range order.MenuList {
		if i == 0 {
			menuText = item.Name
		} else {
			menuText += ", " + item.Name
		}
		priceText += item.Price
	}
	orderDetail := fmt.Sprintf("ご注文内容確認\n\n1. 日程 : %s\n2. 時間 : %s\n3. 発送場所: %s\n3. 品物 : %s\n\n4. お会計: ¥%d",
		order.Date, order.Time, order.Location, menuText, priceText)

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
func makeSorryMessage(cause string) *linebot.TextMessage {
	message := "申し訳ありません。\n" + cause + "\n最初から注文をやり直してください。"
	return linebot.NewTextMessage(message)
}
