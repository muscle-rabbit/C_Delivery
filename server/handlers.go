package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

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
			userID, err := app.fetchUserByLINEProfile(p)
			if err != nil {
				log.Fatal(err)
			}
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if message.Text == "配達員ログイン" {
					ok, err := app.authDeliverer(userID)
					if err != nil {
						g.Error(fmt.Errorf("couldn't auth worker user=%s: %v", userID, err))
						g.Abort()
					}
					if ok {
						if err := app.replyWorkerPanel(event, userID); err != nil {
							g.Error(fmt.Errorf("couldn't return workerpanel: %v", err))
							g.Abort()
						}
					}
					if err := app.replyDenyWorkerLogin(event, userID); err != nil {
						g.Error(fmt.Errorf("couldn't return denyworkerpanel: %v", err))
					}

				}
				if app.sessionStore.sessions[userID] != nil {
					if err := app.reply(event, userID); err != nil {
						g.Error(fmt.Errorf("couldn't start order: %v", err))
					}
				}
				if message.Text == "予約開始" {
					if err := app.reply(event, userID); err != nil {
						g.Error(fmt.Errorf("couldn't start order: %v", err))
					}
				}
			}
		}
	}
}

func (app *app) getUserHandler(g *gin.Context) {
	userID := g.Param("userID")
	userDocument, err := app.fetchUserByDocID(userID)
	if err != nil {
		g.Error(fmt.Errorf("couldn't fetchUser in getUserHandler: %v", err))
	}

	g.JSON(200, &userDocument)
}

func (app *app) getOrdersHandler(g *gin.Context) {
	orderDocuments, err := app.fetchOrderDocuments()
	if err != nil {
		g.Error(fmt.Errorf("couldn't fetchOrders in orderListHandler: %v", err))
	}

	g.JSON(200, &orderDocuments)
	return
}

func (app *app) getOrderHanlder(g *gin.Context) {
	orderID := g.Param("orderID")
	orderDocument, err := app.fetchUserOrder(orderID)
	if err != nil {
		g.Error(fmt.Errorf("coudln't fetchUserOder in getOrderHandler: %v", err))
	}

	g.JSON(200, &orderDocument)
}

func (app *app) changeTradeStatusHandler(g *gin.Context) {
	action := g.Param("action")
	orderID := g.Param("orderID")
	switch action {
	case "/finish":
		if err := app.finishTrade(orderID); err != nil {
			g.Error(fmt.Errorf("couldn't update order in finishTradeHandler: %v", err))
		}

	case "/unfinish":
		if err := app.unfinishTrade(orderID); err != nil {
			g.Error(fmt.Errorf("couldn't update order in finishTradeHandler: %v", err))
		}
	default:
		g.Error(fmt.Errorf("coudln't parse action in finishTradeHandler"))
	}

	orderDoc, err := app.fetchUserOrder(orderID)
	if err != nil {
		g.Error(fmt.Errorf("couldn't fetch order in finishTradeHandler: %v", err))
	}
	g.JSON(200, &orderDoc)
}

func (app *app) postChatHandler(g *gin.Context) {
	var message Message
	chatID := g.Param("chatID")
	err := g.BindJSON(&message)
	message.CreatedAt = time.Now()

	if err != nil {
		g.Error(fmt.Errorf("coudln't parse reader in postChatHandler: %v", err))
	}

	if err = app.postOrderChats(chatID, message); err != nil {
		g.Error(fmt.Errorf("coudln't post order chats in postChatHandler: %v", err))
	}
}

func (app *app) getChatHandler(g *gin.Context) {
	chatID := g.Param("chatID")
	chatDoc, err := app.fetchOrderChats(chatID)
	if err != nil {
		g.Error(fmt.Errorf("coudln't fetch order chats in getChatHandler: %v", err))
		g.Abort()
	}

	g.JSON(200, &chatDoc)
}

func healthcheckHandler(g *gin.Context) {
	g.JSON(200, "ok!")
}
