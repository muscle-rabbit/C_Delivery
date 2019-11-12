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
	"google.golang.org/api/option"
)

// User represents a profile
type User struct {
	UserID      string    `firestore:"user_id,omitempty"`
	DisplayName string    `firestore:"display_name,omitempty"`
	CreatedAt   time.Time `firestore:"created_at,omitempty"`
}

func newApp() (*app, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error get env: %v", err)
	}

	// CREDENTIAL
	fmt.Println("initialize app")
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	client, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

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

	//
	app.service.businessHours = businessHours{today: parseTime(time.Now()), begin: detailTime{12, 00}, end: detailTime{15, 00}, interval: 30, lastorder: "12:30"}

	// linebot の初期化
	if err := app.bot.createBot(); err != nil {
		return nil, err
	}

	return app, nil

}

func (app *app) addUser(profile *linebot.UserProfileResponse) (string, error) {
	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return "", fmt.Errorf("couldn't create client in addUser: %v", err)
	}

	// user がすでに登録されていたら nil を返す。
	iter := client.Collection("users").Where("user_id", "==", profile.UserID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			if doc == nil {
				ref, _, err := client.Collection("users").Add(ctx, User{UserID: profile.UserID, DisplayName: profile.DisplayName, CreatedAt: time.Now()})
				if err != nil {
					return "", fmt.Errorf("couldn't find user document ref: %v", err)
				}
				return ref.ID, nil
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
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't create oreder in fetchUserOrderID: %v", err)
	}

	iter := client.Collection("orders").Documents(ctx)
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

	return orderDocumens, err
}

func (app *app) createOrder(userID string) error {
	userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in addUser: %v", err)
	}

	ref, _, err := client.Collection("orders").Add(ctx, Order{UserID: userID, CreatedAt: time.Now(), Finished: false, InProgress: true})
	if err != nil {
		return fmt.Errorf("couldn't create document in createOrder: %v", err)
	}

	userSession.orderID = ref.ID

	return nil
}

func (app *app) updateOrderInChat(userID string, order Order, prevStep int) error {
	userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in addUser: %v", err)
	}

	switch prevStep {
	case reservateDate:
		_, err = client.Collection("orders").Doc(userSession.orderID).Set(ctx, map[string]interface{}{
			"date": order.Date,
		}, firestore.MergeAll)
		if err != nil {
			return fmt.Errorf("couldn't update order in updateOrder: %v", err)
		}
	case reservateTime:
		if isTimeMessage(order.Time) {
			_, err = client.Collection("orders").Doc(userSession.orderID).Set(ctx, map[string]interface{}{
				"time": order.Time,
			}, firestore.MergeAll)
			if err != nil {
				return fmt.Errorf("couldn't update order in updateOrder: %v", err)
			}
		} else {
			_, err = client.Collection("orders").Doc(userSession.orderID).Set(ctx, map[string]interface{}{
				"products": order.Products,
			}, firestore.MergeAll)
			if err != nil {
				return fmt.Errorf("couldn't update order in updateOrder: %v", err)
			}
		}
	case setLocation:
		_, err = client.Collection("orders").Doc(userSession.orderID).Set(ctx, map[string]interface{}{
			"location": order.Location,
		}, firestore.MergeAll)
	}
	return nil
}

func (app *app) completeOrderInChat(userID string) error {
	userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in completeOrderInChat: %v", err)
	}

	_, err = client.Collection("orders").Doc(userSession.orderID).Set(ctx, map[string]interface{}{
		"in_progress": false,
	}, firestore.MergeAll)

	return err
}

func (app *app) finishTrade(userID string) error {
	userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in finshTrade: %v", err)
	}

	_, err = client.Collection("orders").Doc(userSession.orderID).Set(ctx, map[string]interface{}{
		"finished": true,
	}, firestore.MergeAll)

	return err
}

func (app *app) updateOrderFromDeliveryPanel(orderDocument OrderDocument) error {
	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in addUser: %v", err)
	}

	_, err = client.Collection("orders").Doc(orderDocument.ID).Set(ctx, map[string]interface{}{
		"finished": orderDocument.Order.Finished,
	}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("couldn't update order in updateOrder: %v", err)
	}

	return nil
}

func (app *app) deleteOrder(userID string) error {
	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in addUser: %v", err)
	}

	iter := client.Collection("orders").Where("user_id", "==", userID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil
		}
		if _, err := doc.Ref.Delete(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (app *app) fetchUserOrder(userID string) (Order, error) {
	userSession := app.sessionStore.sessions[userID]

	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return Order{}, fmt.Errorf("couldn't create oreder in fetchUserOrderID: %v", err)
	}

	var order Order
	doc, err := client.Collection("orders").Doc(userSession.orderID).Get(ctx)
	doc.DataTo(&order)
	return order, nil
}

func (app *app) reserve(userID string, products Products) error {
	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in reserveProduct: %v", err)
	}

	batch := client.Batch()

	for id, n := range products {
		var stocks Products
		iter := client.Collection("stocks").Where("product_id", "==", id).Documents(ctx)
		if err != nil {
			return fmt.Errorf("couldn't fetch Product in reserve: %v", err)
		}

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			doc.DataTo(stocks)
			if n >= stocks[id] {
				batch.Set(doc.Ref, stocks[id]-n)
			}
		}
	}

	if _, err = batch.Commit(ctx); err != nil {
		return fmt.Errorf("couldn't commit batch in reserve: %v", err)
	}
	return err
}

func (app *app) fetchProduct(productID string) (Item, error) {
	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return Item{}, fmt.Errorf("couldn't create client in fetchProduct: %v", err)
	}

	doc, err := client.Collection("products").Doc(productID).Get(ctx)
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
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't create client in getMenu: %v", err)
	}

	iter := client.Collection("products").Documents(ctx)
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
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't create client in getLocations: %v", err)
	}

	iter := client.Collection("locations").Documents(ctx)
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

func parseTime(time time.Time) string {
	return time.Format("20060102")
}
