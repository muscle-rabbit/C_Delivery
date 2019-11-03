package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

var c client

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// firestore の初期化
	app, err := newApp()
	if err != nil {
		log.Fatal(err)
	}

	// cookieStore は client のセッション ID をまとめて管理する

	c.bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
	http.HandleFunc("/callback", app.callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func (app *app) callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := c.bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			c.request = r
			c.writer = w
			p, _ := c.bot.GetProfile(event.Source.UserID).Do()
			if err = app.addUser(p); err != nil {
				log.Fatal(err)
			}

			if err != nil {
				log.Fatalf("couldn't create session: %v", err)
			}
			switch event.Message.(type) {
			case *linebot.TextMessage:
				if err := c.reply(event); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
