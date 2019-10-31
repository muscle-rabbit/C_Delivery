package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

var c client
var sessionStore sessions.Store

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// cookieStore は client のセッション ID をまとめて管理する
	cookieStore := sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SECRET")))
	sessionStore = cookieStore

	c.bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
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
			c.session, err = sessionStore.Get(r, event.Source.UserID)
			if err != nil {
				log.Fatalf("couldn't create session: %v", err)
			}
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if err := c.reply(event); err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Message from user is: %s", message.Text)
			}
		}
	}
}
