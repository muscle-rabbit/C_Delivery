package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client
var sessionStore sessions.Store

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// cookieStore は client のセッション ID をまとめて管理する
	cookieStore := sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SECRET")))
	sessionStore = cookieStore

	bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

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
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				quota, err := bot.GetMessageQuota().Do()
				if err := (event, bot); err != nil {
					log.Fatal(err)
				}
				// if _, err = bot.ReplyMessage(event.ReplyToken, messages.ReplyReservationTime(bot)).Do(); err != nil {
				// 	log.Print(err)
				// }
				// if _, err = bot.ReplyMessage(event.ReplyToken, messages.ReplyMenuText(bot)).Do(); err != nil {
				// 	log.Print(err)
				// }
				// if _, err = bot.ReplyMessage(event.ReplyToken, messages.ReplyMenu(bot)).Do(); err != nil {
				// 	log.Print(err)
				// }
				// if _, err = bot.ReplyMessage(event.ReplyToken, messages.ReplyConfirmationText(bot), messages.ReplyConfirmationButton(bot)).Do(); err != nil {
				// 	log.Print(err)
				// }
				// if _, err = bot.ReplyMessage(event.ReplyToken, messages.ReplyThankYou(bot)).Do(); err != nil {
				// 	log.Print(err)
				// }
				if err != nil {
					log.Println("Quota err:", err)
				}

				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.ID+":"+message.Text+" OK! remain message:"+strconv.FormatInt(quota.Value, 10))).Do(); err != nil {
				}
			}
		}
	}
}
