package components

import (
	"github.com/line/line-bot-sdk-go/linebot"
)

func ReplyReservationDate(bot *linebot.Client) *linebot.TemplateMessage {
	title := "日程指定"
	phrase := "ご注文日をお選びください。"
	template := linebot.NewButtonsTemplate(
		"", title, phrase,
		linebot.NewMessageAction("本日 (11/1)", "本日 (11/1)"),
		linebot.NewMessageAction("明日 (11/2)", "本日 (11/2)"),
	)

	message := linebot.NewTemplateMessage("代わりのテキスト", template)
	return message
}
