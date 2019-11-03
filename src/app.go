package main

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/line/line-bot-sdk-go/linebot"
)

type client struct {
	bot     *linebot.Client
	session *sessions.Session
	request *http.Request
	writer  http.ResponseWriter
}
