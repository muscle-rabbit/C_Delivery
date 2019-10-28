package messages

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"time"
)

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
