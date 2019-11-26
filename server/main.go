package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

	r.GET("/healthcheck", healthcheckHandler)

	// linebot のリクエストエンドポイント
	r.POST("/api/v1/callback", app.callbackHandler)

	// 配達員画面からのエンドポイント
	r.GET("/api/v1/order_list", app.getOrdersHandler)

	r.GET("/api/v1/user/:userID", app.getUserHandler)

	// 注文を取得する用のエンドポイント
	r.GET("/api/v1/order/:orderID", app.getOrderHanlder)
	r.GET("/api/v1/order/:orderID/*action", app.changeTradeStatusHandler)

	// チャットエンドポイント
	r.GET("/api/v1/user/:userID/order/:orderID/chats/:chatID", app.getChatHandler)
	r.POST("/api/v1/chats/:chatID", app.postChatHandler)
	r.GET("/api/v1/chats/:chatID", app.getChatHandler)

	// セッションの監視
	go app.watchSessions(time.Second * 3)

	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	fmt.Println("listening... :", addr)
	r.Run(addr)
}
