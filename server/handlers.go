package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

func (bot *Bot) callbackHandler(g *gin.Context) {
	events, err := bot.ParseRequest(g.Request)

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
			p, _ := bot.Client.GetProfile(event.Source.UserID).Do()
			userID, err := app.fetchUserByLINEProfile(p)
			if err != nil {
				log.Fatal(err)
			}
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if message.Text == "配達員ログイン" {
					if ok := authWorker(userID); ok {
						if err := app.replyWorkerPanel(event, userID); err != nil {
							g.Error(fmt.Errorf("couldn't return workerpanel: %v", err))
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

func authWorker(userID string) bool {
	autherID := "aa"
	if userID == autherID {
		return true
	}
	return false
}
