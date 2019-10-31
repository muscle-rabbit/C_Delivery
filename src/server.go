package main

import (
	"github.com/gorilla/sessions"
	"github.com/line/line-bot-sdk-go/linebot"
)

type client struct {
	bot     *linebot.Client
	session *sessions.Session
}
