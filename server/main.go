package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
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

	// gin の生成。静的ファイルサーバー
	r := gin.Default()
	r.LoadHTMLGlob("../dist/*.html")        // load the built dist path
	r.LoadHTMLFiles("static/*/*")           //  load the static path
	r.Static("/static", "../dist/static")   // use the loaded source
	r.StaticFile("/", "../dist/index.html") // use the loaded sourc

	// dev 用ミドルウェア
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:8080"},
		AllowMethods:  []string{"POST", "OPTIONS", "GET"},
		AllowHeaders:  []string{"*"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Minute,
	}))

	// linebot のリクエストエンドポイント
	r.POST("/callback", app.callbackHandler)

	// 配達員画面からのエンドポイント
	r.GET("/order_list", app.getOrderListHandler)
	r.POST("/order", app.postOrderHandler)

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
			user, err := app.fetchUserByLINEProfile(p)
			if err != nil {
				log.Fatal(err)
			}

			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if app.sessionStore.sessions[user.UserID] != nil {
					if err := app.reply(event, user); err != nil {
						log.Fatal(err)
					}
				}
				if message.Text == "予約開始" {
					if err := app.reply(event, user); err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}
}

func (app *app) getOrderListHandler(g *gin.Context) {
	orderDocuments, err := app.fetchOrderDocuments()
	if err != nil {
		g.Error(fmt.Errorf("couldn't fetchOrders in orderListHandler: %v", err))
	}

	g.JSON(200, &orderDocuments)
	return
}

func (app *app) postOrderHandler(g *gin.Context) {
	var orderDocument OrderDocument
	err := g.BindJSON(&orderDocument)
	if err != nil {
		g.Error(fmt.Errorf("coudln't parse reader in postOrderHandler: %v", err))
	}

	if err := app.toggleOrderFinishedStatus(orderDocument); err != nil {
		g.Error(fmt.Errorf("couldn't update order in updateOrderFromDeliveryPanel: %v", err))
	}
}
