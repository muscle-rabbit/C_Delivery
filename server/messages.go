package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
)

type Menu []ProductDocument

type ProductItem struct {
	Name       string `firestore:"name,omitempty"`
	Price      int    `firestore:"price,omitempty"`
	PictureURL string `firestore:"picture_url,omitempty"`
}

// Products は Message API と共有する製品情報です。
type Products map[string]*Product

// Product は ユーザのセッションで保持する製品情報をまとめた構造体です。
type Product struct {
	Name     string `firestore:"name,omitempty" json:"name"`
	Stock    int    `firestore:"stock,omitempty" json:"stock"`
	Reserved bool   `firestore:"reserved,omitempty" json:"reserved"`
}

type ProductDocument struct {
	ID          string
	ProductItem ProductItem
}

func (m Menu) searchProductByName(key string) (ProductDocument, error) {
	for _, doc := range m {
		if doc.ProductItem.Name == key {
			return doc, nil
		}
	}
	// TODO: nil を返したい。
	return ProductDocument{}, fmt.Errorf("This product doesn't exist in Menu: %s", key)
}

func (m Menu) searcProductByID(id string) (ProductDocument, error) {
	for _, doc := range m {
		if doc.ID == id {
			return doc, nil
		}
	}
	return ProductDocument{}, fmt.Errorf("This product doesn't exist in Menu: %", id)
}

func (m Menu) calcPrice(products Products) int {
	var price int
	for id, product := range products {
		for _, v := range m {
			if v.ID == id {
				price += v.ProductItem.Price * product.Stock
			}
		}
	}
	return price
}

var wdays = [...]string{"日", "月", "火", "水", "木", "金", "土"}

type Order struct {
	User       User      `firestore:"user,omitempty" json:"user"`
	Date       string    `firestore:"date,omitempty" json:"date"`
	Time       string    `firestore:"time,omitempty" json:"time"`
	Location   string    `firestore:"location,omitempty" json:"location"`
	Products   Products  `firestore:"products,omitempty" json:"products"`
	TotalPrice int       `firestore:"total_price,omitempty" json:"total_price"`
	InTrade    bool      `firestore:"in_trade,omitempty" json:"in_trade"`
	InProgress bool      `firestore:"in_progress,omitempty" json:"in_progress"`
	CreatedAt  time.Time `firestore:"created_at,omitempty" json:"created_at"`
	UpdatedAt  time.Time `firestore:"updated_at,omitempty" json:"updated_at"`
	ChatID     string    `firestore:"chat_id,omitempty" json:"chat_id"`
}

// type OrderList map[string]Order

type OrderDocument struct {
	ID    string `firestore:"document_id,omitempty" json:"document_id"`
	Order Order  `firestore:"order,omitempty" json:"order"`
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
func (app *app) makeReservationTimeMessage() *linebot.TemplateMessage {
	title := "時間指定"
	phrase := "ご注文時間をお選びください。\nラストオーダー: " + app.service.businessHours.lastorder
	timeTable := app.service.businessHours.makeTimeTable()

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
func (app *app) makeTimeTableMessageAction() []*linebot.MessageAction {
	timeTable := app.service.businessHours.makeTimeTable()
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
func (app *app) makeMenuMessage() (*linebot.FlexMessage, error) {
	var containers []*linebot.BubbleContainer
	containers = append(containers, makeDecideButton())

	stockDocs, err := app.fetchStocks()
	if err != nil {
		return nil, err
	}
	for _, productDoc := range app.service.menu {
		stock, err := stockDocs.searchDocByProductID(productDoc.ID)
		if err != nil {
			return nil, err
		}
		containers = append(containers, makeMenuCard(productDoc.ProductItem, stock.Stock))
	}
	carousel := &linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: containers,
	}

	message := linebot.NewFlexMessage("Menu List", carousel)
	return message, nil
}

func makeOutOfStockMessage() *linebot.TextMessage {
	return linebot.NewTextMessage("申し訳ありません。\nお選びいただいた商品は現在在庫切れです。")
}

func makeUnselectedProductsMessage() *linebot.TextMessage {
	return linebot.NewTextMessage("商品をお選びください。")
}

// makeLocation は 発送先用のメッセージを返すメソッドです。
func (app *app) makeLocationMessage() *linebot.TemplateMessage {
	title := "発送先指定"
	phrase := "発送先を選択下さい。"
	locations := app.service.locations

	template := linebot.NewButtonsTemplate(
		"", title, phrase,
		linebot.NewMessageAction(locations[0].Name, locations[0].Name),
		linebot.NewMessageAction(locations[1].Name, locations[1].Name),
	)
	return linebot.NewTemplateMessage(title, template)
}

// makeConfirmationText は 注文確認テキスト用メッセージを返すメソッドです。
func (app *app) makeConfirmationTextMessage(userID string) (*linebot.TextMessage, error) {
	var menuText string
	menu := app.service.menu
	userSession := app.sessionStore.sessions[userID]
	for id, product := range userSession.products {
		productDoc, err := menu.searcProductByID(id)
		if err != nil {
			return nil, err
		}
		menuText += "・" + productDoc.ProductItem.Name + "x" + strconv.Itoa(product.Stock) + "\n"
	}

	orderDoc, err := app.fetchUserOrder(userSession.orderID)
	if err != nil {
		return nil, err
	}

	order := orderDoc.Order

	orderDetail := fmt.Sprintf("ご注文内容確認\n\n1. 日程 : %s\n2. 時間 : %s\n3. 発送場所: %s\n3. 品物 : \n%s\n4. お会計: ¥%d",
		order.Date, order.Time, order.Location, menuText, menu.calcPrice(userSession.products))

	return linebot.NewTextMessage(orderDetail), nil
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

func (app *app) makeDenyWorkerMessage(userID string) (*linebot.TextMessage, error) {
	user, err := app.fetchUserByDocID(userID)
	if err != nil {
		return nil, err
	}
	return linebot.NewTextMessage(user.DisplayName + "さんは配達員として登録されていないため、ログインできませんでした。"), nil
}

func makeMenuCard(item ProductItem, stock int) *linebot.BubbleContainer {
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
							Text:  "商品名",
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
					Layout:  linebot.FlexBoxLayoutTypeBaseline,
					Spacing: linebot.FlexComponentSpacingTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  "値段",
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
					Layout:  linebot.FlexBoxLayoutTypeBaseline,
					Spacing: linebot.FlexComponentSpacingTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  "在庫",
							Color: "#aaaaaa",
							Size:  linebot.FlexTextSizeTypeSm,
							Flex:  linebot.IntPtr(1),
						},
						&linebot.TextComponent{
							Type:   linebot.FlexComponentTypeText,
							Text:   strconv.Itoa(stock) + "個",
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

func (app *app) makeOrderDetail(userID string) (*linebot.FlexMessage, error) {
	userSession := app.sessionStore.sessions[userID]
	orderDoc, err := app.fetchUserOrder(userSession.orderID)
	if err != nil {
		return nil, err
	}
	order := orderDoc.Order
	text, err := app.service.menu.makeMesssageText(order.Products)
	if err != nil {
		return nil, err
	}

	container := &linebot.BubbleContainer{
		Type: "bubble",
		Body: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeMd,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:    linebot.FlexComponentTypeText,
					Text:    order.User.DisplayName + "さんの\nご注文詳細",
					Wrap:    true,
					Weight:  "bold",
					Gravity: "center",
					Size:    linebot.FlexTextSizeTypeXl,
				},
				&linebot.BoxComponent{
					Type:    linebot.FlexComponentTypeBox,
					Layout:  linebot.FlexBoxLayoutTypeVertical,
					Margin:  "lg",
					Spacing: linebot.FlexComponentSpacingTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.BoxComponent{
							Type:    linebot.FlexComponentTypeBox,
							Layout:  linebot.FlexBoxLayoutTypeBaseline,
							Spacing: linebot.FlexComponentSpacingTypeSm,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  "日程",
									Color: "#aaaaaa",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(1),
								},
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  order.Date,
									Color: "#666666",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(4),
								},
							},
						},
						&linebot.BoxComponent{
							Type:    linebot.FlexComponentTypeBox,
							Layout:  linebot.FlexBoxLayoutTypeBaseline,
							Spacing: linebot.FlexComponentSpacingTypeSm,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  "時間",
									Color: "#aaaaaa",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(1),
								},
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  order.Time,
									Color: "#666666",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(4),
								},
							},
						},
						&linebot.BoxComponent{
							Type:    linebot.FlexComponentTypeBox,
							Layout:  linebot.FlexBoxLayoutTypeBaseline,
							Spacing: linebot.FlexComponentSpacingTypeSm,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  "場所",
									Color: "#aaaaaa",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(1),
								},
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  order.Location,
									Color: "#666666",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(4),
								},
							},
						},
						&linebot.BoxComponent{
							Type:    linebot.FlexComponentTypeBox,
							Layout:  linebot.FlexBoxLayoutTypeBaseline,
							Spacing: linebot.FlexComponentSpacingTypeSm,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  "商品",
									Color: "#aaaaaa",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(1),
								},
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Wrap:  true,
									Text:  text,
									Color: "#666666",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(4),
								},
							},
						},
						&linebot.BoxComponent{
							Type:    linebot.FlexComponentTypeBox,
							Layout:  linebot.FlexBoxLayoutTypeBaseline,
							Spacing: linebot.FlexComponentSpacingTypeSm,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  "合計",
									Color: "#aaaaaa",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(1),
								},
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  "¥ " + strconv.Itoa(order.TotalPrice),
									Color: "#666666",
									Size:  linebot.FlexTextSizeTypeSm,
									Flex:  linebot.IntPtr(4),
								},
							},
						},
					},
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.SpacerComponent{
					Type: linebot.FlexComponentTypeSpacer,
					Size: linebot.FlexSpacerSizeTypeXs,
				},
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Color:  "#905c44",
					Action: linebot.NewURIAction("チャットを開始する。", os.Getenv("HOST_API")+"/#/?user="+userID+"&order="+userSession.orderID+"&chats="+order.ChatID),
				},
			},
		},
	}
	return linebot.NewFlexMessage("注文詳細", container), nil
}

func makeWorkerPanelMessage(userID string) *linebot.FlexMessage {
	container := &linebot.BubbleContainer{
		Type: "bubble",
		Body: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeMd,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:    linebot.FlexComponentTypeText,
					Text:    "配達員専用",
					Wrap:    true,
					Weight:  "bold",
					Gravity: "center",
					Size:    linebot.FlexTextSizeTypeXl,
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.SpacerComponent{
					Type: linebot.FlexComponentTypeSpacer,
					Size: linebot.FlexSpacerSizeTypeXs,
				},
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Color:  "#905c44",
					Action: linebot.NewURIAction("開く", os.Getenv("HOST_API")+"/#/?status=worker&user="+userID),
				},
			},
		},
	}
	return linebot.NewFlexMessage("配達員専用画面", container)
}

func makeAskCorrectMessages() *linebot.TextMessage {
	return linebot.NewTextMessage("表示された項目からお選びください。")
}
