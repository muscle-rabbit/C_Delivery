package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
)

type Item struct {
	ID         string
	Name       string `firestore:"name,omitempty"`
	Price      int    `firestore:"price,omitempty"`
	PictureURL string `firestore:"picture_url,omitempty"`
}

type Menu []Item

func (m Menu) searchItemByName(key string) Item {
	for _, item := range m {
		if item.Name == key {
			return item
		}
	}
	// TODO: nil を返したい。
	return Item{}
}

func (m Menu) searchItemIDByName(key string) string {
	for _, item := range m {
		if item.Name == key {
			return item.ID
		}
	}
	return ""
}

func (m Menu) searchItemNameByID(id string) string {
	for _, item := range m {
		if item.ID == id {
			return item.Name
		}
	}
	return ""
}

var wdays = [...]string{"日", "月", "火", "水", "木", "金", "土"}

type Order struct {
	UserID    string    `firestore:"user_id,omitempty"`
	Date      string    `firestore:"date,omitempty"`
	Time      string    `firestore:"time,omitempty"`
	Location  string    `firestore:"location,omitempty"`
	Products  Products  `firestore:"products,omitempty"`
	CreatedAt time.Time `firestore:"created_at,omitempty"`
	UpdatedAt time.Time `firestore:"updated_at,omitempty"`
}

func (bh businessHours) makeTimeTable() []string {
	today := time.Now()
	jst, _ := time.LoadLocation("Asia/Tokyo")

	begin := time.Date(today.Year(), today.Month(), today.Day(), bh.begin.hour, bh.begin.minute, 0, 0, jst)
	end := time.Date(today.Year(), today.Month(), today.Day(), bh.end.hour, bh.end.minute, 0, 0, jst)

	diff := end.Sub(begin)
	interval := time.Duration(bh.interval) * time.Minute

	n := int(diff / interval)
	tt := make([]string, n+1)
	layout := "3:04AM"

	for i := 0; i < n+1; i++ {
		if i == 0 {
			tt[0] = begin.Format(layout) + "~" + begin.Add(interval).Format(layout)
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

func makeMenuMessage(menu Menu) *linebot.FlexMessage {
	var containers []*linebot.BubbleContainer
	containers = append(containers, makeDecideButton())
	for _, item := range menu {
		containers = append(containers, makeMenuCard(item))
	}
	carousel := &linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: containers,
	}

	message := linebot.NewFlexMessage("Menu List", carousel)
	return message
}

func makeHalfConfirmation(products Products, menu Menu, order Order, price int) *linebot.TextMessage {
	var menuText string

	for id, n := range products {
		menuText += "・" + menu.searchItemNameByID(id) + " x " + strconv.Itoa(n) + "\n"
	}
	return linebot.NewTextMessage(fmt.Sprintf("ご注文途中確認\n\nお間違いなければ次のステップに移ります。\n\n1. 日程 : %s\n2. 時間 : %s\n3. 品物: \n%s\n\n4. お会計 : ¥%d", order.Date, order.Time, menuText, price))
}

// makeLocation は 発送先用のメッセージを返すメソッドです。
func makeLocationMessage(locations []Location) *linebot.TemplateMessage {
	title := "発送先指定"
	phrase := "発送先を選択下さい。"

	template := linebot.NewButtonsTemplate(
		"", title, phrase,
		linebot.NewMessageAction(locations[0].Name, locations[0].Name),
		linebot.NewMessageAction(locations[1].Name, locations[1].Name),
	)
	return linebot.NewTemplateMessage(title, template)
}

// makeConfirmationText は 注文確認テキスト用メッセージを返すメソッドです。
func makeConfirmationTextMessage(products Products, menu Menu, order Order, price int) *linebot.TextMessage {
	var menuText string

	for id, n := range products {
		menuText += "・" + menu.searchItemNameByID(id) + " x " + strconv.Itoa(n) + "\n"
	}

	orderDetail := fmt.Sprintf("ご注文内容確認\n\n1. 日程 : %s\n2. 時間 : %s\n3. 発送場所: %s\n3. 品物 : \n%s\n4. お会計: ¥%d",
		order.Date, order.Time, order.Location, menuText, price)

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

func makeMenuCard(item Item) *linebot.BubbleContainer {
	return &linebot.BubbleContainer{
		Type: "bubble",
		Hero: &linebot.ImageComponent{
			Type:        linebot.FlexComponentTypeImage,
			URL:         item.PictureURL,
			Size:        linebot.FlexImageSizeTypeFull,
			AspectRatio: linebot.FlexImageAspectRatioType20to13,
			AspectMode:  linebot.FlexImageAspectModeTypeCover,
		},
		Body: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeMd,
			Contents: []linebot.FlexComponent{
				&linebot.BoxComponent{
					Type:    linebot.FlexComponentTypeBox,
					Layout:  linebot.FlexBoxLayoutTypeBaseline,
					Spacing: linebot.FlexComponentSpacingTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  "Name",
							Color: "#aaaaaa",
							Size:  linebot.FlexTextSizeTypeSm,
							Flex:  linebot.IntPtr(1),
						},
						&linebot.TextComponent{
							Type:   linebot.FlexComponentTypeText,
							Text:   item.Name,
							Weight: linebot.FlexTextWeightTypeBold,
							Size:   linebot.FlexTextSizeTypeMd,
							Flex:   linebot.IntPtr(5),
						},
					},
				},
				&linebot.BoxComponent{
					Type:    linebot.FlexComponentTypeBox,
					Layout:  linebot.FlexBoxLayoutTypeVertical,
					Spacing: linebot.FlexComponentSpacingTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  "Price",
							Color: "#aaaaaa",
							Size:  linebot.FlexTextSizeTypeSm,
							Flex:  linebot.IntPtr(1),
						},
						&linebot.TextComponent{
							Type:   linebot.FlexComponentTypeText,
							Text:   "¥" + strconv.Itoa(item.Price),
							Weight: linebot.FlexTextWeightTypeBold,
							Size:   linebot.FlexTextSizeTypeMd,
							Flex:   linebot.IntPtr(5),
						},
					},
				},
				&linebot.BoxComponent{
					Type:    linebot.FlexComponentTypeBox,
					Layout:  linebot.FlexBoxLayoutTypeVertical,
					Spacing: linebot.FlexComponentSpacingTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  "注文したい個数をお選びください。",
							Align: linebot.FlexComponentAlignTypeCenter,
							Color: "#979A99",
						},
					},
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeHorizontal,
			Spacing: linebot.FlexComponentSpacingTypeMd,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Color:  "#009944",
					Action: linebot.NewMessageAction("x 1", item.Name+"x1"),
				},
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Color:  "#009944",
					Action: linebot.NewMessageAction("x 2", item.Name+"x2"),
				},
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Color:  "#009944",
					Action: linebot.NewMessageAction("x 3", item.Name+"x3"),
				},
			},
		},
	}
}

func makeDecideButton() *linebot.BubbleContainer {

	return &linebot.BubbleContainer{
		Type: "bubble",
		Hero: &linebot.ImageComponent{
			Type:        linebot.FlexComponentTypeImage,
			URL:         "https://mitkp.com/wp-content/uploads/2017/04/pop_kettei.png",
			Size:        linebot.FlexImageSizeTypeFull,
			AspectRatio: linebot.FlexImageAspectRatioType20to13,
			AspectMode:  linebot.FlexImageAspectModeTypeCover,
		},
		Body: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeMd,
			Contents: []linebot.FlexComponent{
				&linebot.BoxComponent{
					Type:    linebot.FlexComponentTypeBox,
					Layout:  linebot.FlexBoxLayoutTypeBaseline,
					Spacing: linebot.FlexComponentSpacingTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type:   linebot.FlexComponentTypeText,
							Text:   "次に進む",
							Weight: linebot.FlexTextWeightTypeBold,
							Size:   linebot.FlexTextSizeTypeMd,
							Flex:   linebot.IntPtr(5),
						},
					},
				},
				&linebot.BoxComponent{
					Type:    linebot.FlexComponentTypeBox,
					Layout:  linebot.FlexBoxLayoutTypeVertical,
					Spacing: linebot.FlexComponentSpacingTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  "注文が決まりましたら押してください。",
							Align: linebot.FlexComponentAlignTypeCenter,
							Color: "#979A99",
						},
					},
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeHorizontal,
			Spacing: linebot.FlexComponentSpacingTypeMd,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Color:  "#009944",
					Action: linebot.NewMessageAction("注文決定", "注文決定"),
				},
			},
		},
	}
}

func (m Menu) calcPrice(products Products) int {
	var price int
	for id, n := range products {
		for _, v := range m {
			if v.ID == id {
				price += v.Price * n
			}
		}
	}
	return price
}