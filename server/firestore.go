package main

import (
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// User represents a profile
type User struct {
	UserID      string    `firestore:"user_id,omitempty" json:"user_id"`
	DisplayName string    `firestore:"display_name,omitempty" json:"display_name"`
	CreatedAt   time.Time `firestore:"created_at,omitempty" json:"-"`
}

type StockDocuments map[string]Stock

type Stock struct {
	ProductID string `firestore:"product_id,omitempty"`
	Stock     int    `firestore:"stock,omitempty"`
}

type ChatsDoc struct {
	Messages []Message `firestore:"messages,omitempty" json:"messages"`
	OrderID  string    `firestore:"order_id,omitempty" json:"order_id"`
}

type Message struct {
	Content   string    `firestore:"content,omitempty" json:"content"`
	CreatedAt time.Time `firestore:"created_at,omitempty" json:"created_at"`
	UserID    string    `firestore:"user_id,omitempty" json:"user_id"`
}

func (sd StockDocuments) searchDocByProductID(productID string) (Stock, error) {
	for docID, stock := range sd {
		if stock.ProductID == productID {
			return sd[docID], nil
		}
	}
	return Stock{}, fmt.Errorf("couldn't find stock doc by: %s", productID)
}

func (sd StockDocuments) searchDocIDByProductID(productID string) (string, error) {
	for docID, stock := range sd {
		if stock.ProductID == productID {
			return docID, nil
		}
	}
	return "", fmt.Errorf("couldn't find stock doc by: %s", productID)
}

func newApp() (*app, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error get env: %v", err)
	}

	// CREDENTIAL
	fmt.Println("initialize app")
	ctx := context.Background()

	conf := &firebase.Config{ProjectID: os.Getenv("PROJECT_ID")}
	application, err := firebase.NewApp(ctx, conf)

	client, err := application.Firestore(ctx)

	app := &app{
		bot:          &bot{},
		client:       client,
		sessionStore: &sessionStore{lifespan: time.Minute * 10, sessions: make(sessions)},
		service:      &service{},
	}

	// 商品情報の取得。
	menu, err := app.getMenu()
	if err != nil {
		return nil, fmt.Errorf("couldn't get Menu : %v", err)
	}
	app.service.menu = menu

	// 配達場所情報の取得。
	locations, err := app.getLocations()
	if err != nil {
		return nil, fmt.Errorf("couldn't get Locations: %v", err)
	}
	app.service.locations = locations

	// サービス時間に関わる情報を取得。
	app.service.businessHours = businessHours{today: parseTime(time.Now()), begin: detailTime{12, 00}, end: detailTime{15, 00}, interval: 30, lastorder: "12:30"}

	// linebot の初期化
	if err := app.bot.createBot(); err != nil {
		return nil, err
	}

	return app, nil

}

func (app *app) addUser(profile *linebot.UserProfileResponse) (string, error) {
	ctx := context.Background()

	ref, _, err := app.client.Collection("users").Add(ctx, User{UserID: profile.UserID, DisplayName: profile.DisplayName, CreatedAt: time.Now()})
	if err != nil {
		return "", fmt.Errorf("couldn't find user document ref: %v", err)
	}

	return ref.ID, nil
}

func (app *app) fetchUserByDocID(docID string) (User, error) {
	var user User
	ctx := context.Background()

	doc, err := app.client.Collection("users").Doc(docID).Get(ctx)
	if err != nil {
		return User{}, fmt.Errorf("couldn't find user document ref: %v", err)
	}

	if err := doc.DataTo(&user); err != nil {
		return User{}, err
	}
	// raw の user ID は LINE ID なので Doc ID に入れ替える。
	user.UserID = doc.Ref.ID
	return user, nil
}

func (app *app) authDeliverer(userID string) (bool, error) {
	ctx := context.Background()
	iter := app.client.Collection("deliverers").Where("user_id", "==", userID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			if doc == nil {
				// 配達員がマッチしなかったら false を返す。
				return false, nil
			}
			break
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (app *app) fetchUserByLINEProfile(profile *linebot.UserProfileResponse) (string, error) {
	ctx := context.Background()

	iter := app.client.Collection("users").Where("user_id", "==", profile.UserID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			if doc == nil {
				// user がいなければ作成。
				userID, err := app.addUser(profile)
				if err != nil {
					return "", err
				}
				return userID, nil
			}
			break
		}
		if err != nil {
			return "", err
		}
		return doc.Ref.ID, nil
	}

	return "", nil
}

func (app *app) fetchOrderDocuments() ([]OrderDocument, error) {
	var orderDocumens []OrderDocument
	ctx := context.Background()

	iter := app.client.Collection("orders").Documents(ctx)
	for {
		var order Order
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		err = doc.DataTo(&order)
		if err != nil {
			return nil, err
		}
		orderDocumens = append(orderDocumens, OrderDocument{ID: doc.Ref.ID, Order: order})
	}

	return orderDocumens, nil
}

func (app *app) createOrder(userID string) error {
	userSession := app.sessionStore.sessions[userID]
	ctx := context.Background()

	user, err := app.fetchUserByDocID(userID)
	if err != nil {
		return err
	}

	ref, _, err := app.client.Collection("orders").Add(ctx, Order{User: user, CreatedAt: time.Now(), InTrade: true, InProgress: true})
	if err != nil {
		return fmt.Errorf("couldn't create document in createOrder: %v", err)
	}

	userSession.orderID = ref.ID

	return nil
}

func (app *app) updateOrderInChat(userID string, order Order) error {
	userSession := app.sessionStore.sessions[userID]
	prevStep := userSession.prevStep

	ctx := context.Background()

	var willUpdated string

	switch prevStep {
	case reservateDate:
		willUpdated = "date"
	case reservateTime:
		if isTimeMessage(order.Time) {
			willUpdated = "time"
		} else {
			willUpdated = "products"
		}
	case setLocation:
		willUpdated = "location"
	case setMenu:
		willUpdated = "total_price"
	}

	if _, err := app.client.Collection("orders").Doc(userSession.orderID).Set(ctx, order, firestore.Merge([]string{willUpdated})); err != nil {
		return fmt.Errorf("couldn't update order in updateOrderInChat: %v", err)
	}
	return nil
}

func (app *app) completeOrderInChat(userID string) error {
	userSession := app.sessionStore.sessions[userID]
	ctx := context.Background()

	chatID, err := app.createOrderChat(userSession.orderID, userID)
	if err != nil {
		return err
	}

	_, err = app.client.Collection("orders").Doc(userSession.orderID).Set(ctx, map[string]interface{}{
		"chat_id": chatID, "in_progress": false, "total_price": app.service.menu.calcPrice(userSession.products)}, firestore.MergeAll)
	return err
}

func (app *app) finishTrade(orderID string) error {
	ctx := context.Background()
	_, err := app.client.Collection("orders").Doc(orderID).Set(ctx, Order{InTrade: false}, firestore.Merge([]string{"in_trade"}))

	return err
}

func (app *app) unfinishTrade(orderID string) error {
	ctx := context.Background()
	_, err := app.client.Collection("orders").Doc(orderID).Set(ctx, Order{InTrade: true}, firestore.Merge([]string{"in_trade"}))

	return err
}

func (app *app) deleteOrder(orderID string) error {
	ctx := context.Background()

	ref := app.client.Collection("orders").Doc(orderID)
	_, err := ref.Delete(ctx)
	return err
}

func (app *app) fetchUserOrder(orderID string) (OrderDocument, error) {
	ctx := context.Background()

	var order Order
	doc, err := app.client.Collection("orders").Doc(orderID).Get(ctx)
	if err != nil {
		return OrderDocument{}, err
	}
	doc.DataTo(&order)
	return OrderDocument{ID: doc.Ref.ID, Order: order}, nil
}

// 在庫切れの場合 true, nil を返す。
func (app *app) reserveProducts(userID string, message string) (bool, error) {
	productM, err := parseMessageToProductText(message, app.service.menu)
	if err != nil {
		return false, err
	}
	userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()

	stockDocs, err := app.fetchStocks()
	if err != nil {
		return false, err
	}

	for id, product := range productM {
		if product.Reserved {
			continue
		}
		stock, err := stockDocs.searchDocByProductID(id)
		if err != nil {
			return false, err
		}
		if stock.Stock >= product.Stock {
			stock.Stock -= product.Stock
			docID, err := stockDocs.searchDocIDByProductID(id)
			if err != nil {
				return false, err
			}
			if _, err := app.client.Collection("stocks").Doc(docID).Set(ctx, stock); err != nil {
				return false, err
			}
			if err := userSession.products.setProduct(productM); err != nil {
				return false, err
			}
			userSession.products[id].Reserved = true
			return false, nil
		}
		return true, nil
	}
	return false, err
}

func (app *app) restoreStocks(userID string) error {
	userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()

	stockDocs, err := app.fetchStocks()
	if err != nil {
		return err
	}

	for id, product := range userSession.products {
		stock, err := stockDocs.searchDocByProductID(id)
		if err != nil {
			return err
		}
		stock.Stock += product.Stock
		docID, err := stockDocs.searchDocIDByProductID(id)
		if err != nil {
			return err
		}
		if _, err := app.client.Collection("stocks").Doc(docID).Set(ctx, stock); err != nil {
			return err
		}
	}

	return nil
}

func (app *app) fetchStocks() (StockDocuments, error) {
	stockDocs := make(StockDocuments)
	ctx := context.Background()

	iter := app.client.Collection("stocks").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var stock Stock
		if err := doc.DataTo(&stock); err != nil {
			return nil, err
		}
		stockDocs[doc.Ref.ID] = stock
	}
	return stockDocs, nil
}

func (app *app) fetchProduct(productID string) (Item, error) {
	ctx := context.Background()

	doc, err := app.client.Collection("products").Doc(productID).Get(ctx)
	if err != nil {
		return Item{}, fmt.Errorf("couldn't get document in fetchProduct: %v", err)
	}

	var item Item
	if err := doc.DataTo(&item); err != nil {
		return Item{}, err
	}

	return item, err
}

func (app *app) getMenu() (Menu, error) {
	var menu Menu
	ctx := context.Background()

	iter := app.client.Collection("products").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var item Item
		if err := doc.DataTo(&item); err != nil {
			return nil, err
		}
		item.ID = doc.Ref.ID
		menu = append(menu, item)
	}

	return menu, nil
}

func (app *app) getLocations() ([]Location, error) {
	var locations []Location
	ctx := context.Background()

	iter := app.client.Collection("locations").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var location Location
		if err := doc.DataTo(&location); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}

	return locations, nil
}

func (app *app) cancelOrder(userID string) error {
	userSession := app.sessionStore.sessions[userID]
	if err := app.deleteOrder(userSession.orderID); err != nil {
		return err
	}
	if err := app.restoreStocks(userID); err != nil {
		return err
	}
	if err := app.sessionStore.deleteUserSession(userID); err != nil {
		return err
	}
	return nil
}

func (app *app) fetchOrderChats(chatID string) (ChatsDoc, error) {
	var chatDoc ChatsDoc
	ctx := context.Background()

	doc, err := app.client.Collection("chats").Doc(chatID).Get(ctx)
	if err != nil {
		return ChatsDoc{}, err
	}
	if err = doc.DataTo(&chatDoc); err != nil {
		return ChatsDoc{}, err
	}
	return chatDoc, nil
}

func (app *app) postOrderChats(chatID string, message Message) error {
	ctx := context.Background()

	chat := app.client.Collection("chats").Doc(chatID)
	_, err := chat.Update(ctx, []firestore.Update{
		{Path: "messages", Value: firestore.ArrayUnion(message)},
	})
	if err != nil {
		return err
	}
	return err
}

func (app *app) createOrderChat(orderID string, userID string) (string, error) {
	ctx := context.Background()

	ref, _, err := app.client.Collection("chats").Add(ctx, map[string]interface{}{
		"order_id": orderID,
	})
	if err != nil {
		return "", err
	}
	return ref.ID, err
}

func parseTime(time time.Time) string {
	return time.Format("20060102")
}
