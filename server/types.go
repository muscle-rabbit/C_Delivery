package main

import (
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/shinyamizuno1008/C_Delivery/server/firestore"
)

type Client struct {
	Bot          Bot
	SessionStore *SessionStore
}
type Bot struct {
	Client *linebot.Client
	Menu   firestore.Menu
}

const (
	Begin int = iota
	ReservateDate
	ReservateTime
	SetMenu
	SetLocation
	Confirm
	End
)
