package main

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"github.com/line/line-bot-sdk-go/linebot"
)

// User represents a profile
type User struct {
	UserID      string `firestore:"profileid,omitempty"`
	DisplayName string `firestore:"display_name,omitempty"`
	PictureURL  string `firestore:"picture_url,omitempty"`
}

type App struct {
	*firebase.App
}

func (app App) addUser(profile *linebot.UserProfileResponse) error {
	ctx := context.Background()
	client, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("couldn't create client in addUser: %v", err)
	}

	// Todo: user がすでに登録されていれば 即時 return
	iter := client.Collection("users").Where("user_id", "==", profile.UserID).Documents(ctx)
	doc, err := iter.Next()
	if err != nil {
		return err
	}
	fmt.Println(doc.Data())

	user := User{UserID: profile.UserID, DisplayName: profile.DisplayName, PictureURL: profile.PictureURL}

	_, _, err = client.Collection("profile").Add(ctx, user)

	if err != nil {
		return fmt.Errorf("Couldn't add data in addUser: %v", err)
	}
	return nil
}

// TODO: 注文を追加するメソッドの追加 Create
func (app App) addOrder() error {
	var err error
	return err
}

// TODO: 注文を取得するメソッドの追加 Read
func (app App) getOrder() error {
	var err error
	return err
}

// TODO: 注文を更新するメソッドの追加 Update
func (app App) editOrder() error {
	var err error
	return err
}

// TODO: 注文を削除するメソッドの追加 Delete
func (app App) deleteOrder() error {
	var err error
	return err
}
