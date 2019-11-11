package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// app の初期化
	app, err := newApp()
	if err != nil {
		log.Fatal(err)
	}

	// gin の生成。
	r := gin.Default()
	r.LoadHTMLGlob("../dist/*.html")        // load the built dist path
	r.LoadHTMLFiles("static/*/*")           //  load the static path
	r.Static("/static", "../dist/static")   // use the loaded source
	r.StaticFile("/", "../dist/index.html") // use the loaded sourc
	r.POST("/callback", app.callbackHandler)

	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	r.Run(addr)
}

func (app *app) callbackHandler(g *gin.Context) {
	events, err := app.bot.client.ParseRequest(g.Request)

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
			p, _ := app.bot.client.GetProfile(event.Source.UserID).Do()
			docID, err := app.addUser(p)
			if err != nil {
				log.Fatal(err)
			}

			if err != nil {
				log.Fatalf("couldn't create session: %v", err)
			}
			switch event.Message.(type) {
			case *linebot.TextMessage:
				if err := app.reply(event, docID); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
