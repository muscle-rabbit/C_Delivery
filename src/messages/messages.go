package messages

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"strconv"
	"time"
)

type Item struct {
	Name     string
	Price    int
	ImageURL string
}

// ReplyReservationDate は 注文日程指定用のメッセージを返すメソッドです。
func ReplyReservationDate(bot *linebot.Client) *linebot.TemplateMessage {
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

// ReplyReservationTime は 注文時間指定用のメッセージを返すメソッドです。
func ReplyReservationTime(bot *linebot.Client) *linebot.TemplateMessage {
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

// ReplyMenu は 商品指定用のメッセージを返すメソッドです。
func ReplyMenu(bot *linebot.Client) *linebot.TemplateMessage {
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
	log.Println("in message before return.")
	return linebot.NewTemplateMessage("メニュー指定", template)
}

func ReplyLocation(bot *linebot.Client) *linebot.TemplateMessage {
	locations := []string{"8号間 2F 中央広場"}
	return linebot.NewTemplateMessage("日程指定", template)
}
