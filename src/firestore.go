package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"

	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"

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
		client:       client,
		sessionStore: &sessionStore{lifespan: time.Minute * 10, sessions: make(sessions)},
		service:      &service{},
	}

	// 商品情報の取得。
	menu, err := app.getMenu()
	if err != nil {
		return nil, fmt.Errorf("couldn't get Menu : %v", err)
	}
	app.service.menu = &menu

	// 配達場所情報の取得。

	locations, err := app.getLocations()
	if err != nil {
		return nil, fmt.Errorf("couldn't get Locations: %v", err)
	}

	app.service.locations = locations
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

// TODO: 注文を追加するメソッドの追加 Create
func (app *app) addOrder() error {
	var err error
	return err
}

// TODO: 注文を取得するメソッドの追加 Read
func (app *app) getOrder() error {
	var err error
	return err
}

// TODO: 注文を更新するメソッドの追加 Update
func (app *app) editOrder() error {
	var err error
	return err
}

// TODO: 注文を削除するメソッドの追加 Delete
func (app *app) deleteOrder() error {
	var err error
	return err
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
		var item item
		if err := doc.DataTo(&item); err != nil {
			return nil, err
		}
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
