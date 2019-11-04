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
	UserID      string `firestore:"user_id,omitempty"`
	DisplayName string `firestore:"display_name,omitempty"`
	CreatedAt   string `firestore:"created_at,omitempty"`
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

	return &app{
		client: client,
		sessionStore: &sessionStore{
			lifespan: time.Minute * 10,
			sessions: make(sessions),
		},
	}, nil
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
	ctx := context.Background()
	client, err := app.client.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in addUser: %v", err)
	}
	fmt.Println(client)

	// user がすでに登録されていたら nil を返す。
	// doc, err := iter.Next()
	// if err != nil {
	// 	if doc == nil {
	// 		user := User{UserID: profile.UserID, DisplayName: profile.DisplayName, CreatedAt: time.Now()}
	// 		ref, _, err := client.Collection("orders").Add(ctx, user)
	// 		if err != nil {
	// 			return "", fmt.Errorf("couldn't find user document ref: %v", err)
	// 		}
	// 		return ref.ID, nil
	// 	} else {
	// 		return "", fmt.Errorf("couldn't iterate user document: %v", err)
	// 	}
	// }

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
