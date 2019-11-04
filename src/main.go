package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

var c bot

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
	app.bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))

	// gin の生成。
	r := gin.Default()
	r.POST("/callback", app.callbackHandler)

	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	r.Run(addr)
}

func (app *app) callbackHandler(g *gin.Context) {
	events, err := app.bot.ParseRequest(g.Request)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			g.Writer.WriteHeader(400)
		} else {
			g.Writer.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			c.request = g.Request
			c.writer = g.Writer
			p, _ := app.bot.GetProfile(event.Source.UserID).Do()
			if err = app.addUser(p); err != nil {
				log.Fatal(err)
			}

			if err != nil {
				log.Fatalf("couldn't create session: %v", err)
			}
			switch event.Message.(type) {
			case *linebot.TextMessage:
				if err := app.reply(event); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
