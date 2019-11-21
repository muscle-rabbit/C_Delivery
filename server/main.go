package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
	r.Static("/js", "../dist/js")           // use the loaded source
	r.Static("/css", "../dist/css")         // use the loaded source
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
	r.GET("/order_list", app.getOrdersHandler)

	r.GET("/user/:userID", app.getUserHandler)

	// 注文を取得する用のエンドポイント
	r.GET("/order/:orderID", app.getOrderHanlder)
	r.GET("/order/:orderID/*action", app.changeTradeStatusHandler)

	// チャットエンドポイント
	r.GET("/user/:userID/order/:orderID/chats/:chatID", app.getChatHandler)
	r.POST("/chats/:chatID", app.postChatHandler)
	r.GET("/chats/:chatID", app.getChatHandler)

	// セッションの監視
	go app.watchSessions(time.Second * 3)

	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	r.Run(addr)
}
