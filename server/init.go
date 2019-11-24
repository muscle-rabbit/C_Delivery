package main

import (
	"context"
	"fmt"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
	"google.golang.org/api/option"
)

func newApp() (*app, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error get env: %v", err)
	}

	// CREDENTIAL
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

	// サービス時間に関わる情報を取得。
	app.service.businessHours = businessHours{today: parseTime(time.Now()), begin: detailTime{12, 00}, end: detailTime{15, 00}, interval: 30, lastorder: "12:30"}

	// linebot の初期化
	if err := app.bot.createBot(); err != nil {
		return nil, err
	}

	return app, nil

}

func (bot *Bot) CreateBot() error {
	var err error
	bot.Client, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
	if err != nil {
		return err
	}
	return err
}
