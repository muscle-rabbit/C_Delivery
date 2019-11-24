package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/shinyamizuno1008/C_Delivery/server/firestore"
)

var wdays = [...]string{"日", "月", "火", "水", "木", "金", "土"}

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

// makeMenuText は 商品指定用のテキストメッセージを返すメソッドです。
func makeMenuTextMessage() *linebot.TextMessage {
	return linebot.NewTextMessage("ご注文品をお選びください。")
}

// makeMenu は 商品指定用の写真付きカルーセルを返すメソッドです。
func makeMenuMessage(menu firestore.Menu, stockDocs firestore.StockDocuments) (*linebot.FlexMessage, error) {
	var containers []*linebot.BubbleContainer
	containers = append(containers, makeDecideButton())

	for _, item := range menu {
		stock, err := stockDocs.SearchDocByProductID(item.ID)
		if err != nil {
			return nil, err
		}
		containers = append(containers, makeMenuCard(item, stock.Stock))
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

func (client *Client) makeHalfConfirmation(order firestore.Order, userID string) (*linebot.TextMessage, error) {
	userSession := client.SessionStore.Sessions[userID]

	if err := app.updateOrderInChat(userID, Order{TotalPrice: app.service.menu.calcPrice(userSession.products)}); err != nil {
		return nil, err
	}

	orderDoc, err := app.fetchUserOrder(userSession.orderID)
	if err != nil {
		return nil, err
	}

	menuText := app.service.menu.makeMesssageText(userSession.products)
	order := orderDoc.Order

	return linebot.NewTextMessage(fmt.Sprintf("ご注文途中確認\n\nお間違いなければ次のステップに移ります。\n\n1. 日程 : %s\n2. 時間 : %s\n3. 品物: \n%s\n\n4. お会計 : ¥%d", order.Date, order.Time, menuText, app.service.menu.calcPrice(userSession.products))), nil
}

// makeLocation は 発送先用のメッセージを返すメソッドです。
func makeLocationMessage(location firestore.Location) *linebot.TemplateMessage {
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
func makeConfirmationTextMessage(userID string) (*linebot.TextMessage, error) {
	var menuText string
	menu := app.service.menu
	userSession := app.sessionStore.sessions[userID]
	for id, product := range userSession.products {
		menuText += "・" + menu.searchItemNameByID(id) + " x " + strconv.Itoa(product.Stock) + "\n"
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

func makeDenyWorkerMessage(userID string) (*linebot.TextMessage, error) {
	user, err := app.fetchUserByDocID(userID)
	if err != nil {
		return nil, err
	}
	return linebot.NewTextMessage(user.DisplayName + "さんは配達員として登録されていないため、ログインできませんでした。"), nil
}

func makeMenuCard(item common.ProductItem, stock int) *linebot.BubbleContainer {
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

func makeOrderDetail(userID string) (*linebot.FlexMessage, error) {
	userSession := app.sessionStore.sessions[userID]
	orderDoc, err := app.fetchUserOrder(userSession.orderID)
	if err != nil {
		return nil, err
	}
	order := orderDoc.Order
	fmt.Println("url: ", os.Getenv("LIFF_ENDPOINT")+"/user/"+userID+"/order/"+userSession.orderID+"/chats/"+order.ChatID)

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
									Text:  app.service.menu.makeMesssageText(order.Products),
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
					Action: linebot.NewURIAction("チャットを開始する。", os.Getenv("LIFF_ENDPOINT")+"/user/"+userID+"/order/"+userSession.orderID+"/chats/"+order.ChatID),
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
					Action: linebot.NewURIAction("開く", os.Getenv("LIFF_ENDPOINT")+"/user/"+userID+"/worker_panel"),
				},
			},
		},
	}
	return linebot.NewFlexMessage("注文詳細", container)
}
