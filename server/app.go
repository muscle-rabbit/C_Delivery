package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/line/line-bot-sdk-go/linebot"
)

type app struct {
	bot          *bot
	client       *firestore.Client
	sessionStore *sessionStore
	service      *service
}

type sessionStore struct {
	sessions sessions
	lifespan time.Duration
}

type bot struct {
	client *linebot.Client
}

type userSession struct {
	orderID   string
	prevStep  int
	createdAt time.Time
	products  Products
}

type sessions map[string]*userSession

type service struct {
	menu          Menu
	locations     []Location
	businessHours businessHours
	detailTime    detailTime
}

type Location struct {
	Name string `firestore:"name,omitempty"`
}

type businessHours struct {
	today     string
	begin     detailTime `firestore:"begin,omitempty"`
	end       detailTime `firestore:"end,omitempty"`
	interval  int        `firestore:"interval,omitempty"`
	lastorder string     `firestore:"lastorder,omitempty"`
}

type detailTime struct {
	hour   int
	minute int
}

// map[{products Document の ID}] 個数
// type sessionOrderedProducts map[string]int

// map[{products Document の ID}] 製品情報
type Products map[string]*Product

func (products Products) setProduct(p Products) error {
	for id, product := range p {
		if products[id] != nil {
			products[id].Stock += product.Stock
			return nil
		}
		products[id] = product
		return nil
	}
	return fmt.Errorf("couldn't set prodct in session")
}

func (menu Menu) makeMesssageText(p Products) string {
	var menuText string
	for id, product := range p {
		menuText += "・" + menu.searchItemNameByID(id) + " x " + strconv.Itoa(product.Stock) + "\n"
	}
	return menuText
}

type Product struct {
	Name     string `firestore:"name,omitempty" json:"name"`
	Stock    int    `firestore:"stock,omitempty" json:"stock"`
	Reserved bool   `firestore:"reserved,omitempty" json:"reserved"`
}

func (ss *sessionStore) createSession(userID string) *userSession {
	ss.sessions[userID] = &userSession{prevStep: begin, createdAt: time.Now(), products: make(Products)}
	return ss.sessions[userID]
}

func (ss *sessionStore) deleteUserSession(userID string) error {
	if ss.sessions[userID] == nil {
		return fmt.Errorf("User doesn't exist in session Store: ID. %v", userID)
	}
	delete(ss.sessions, userID)
	return nil
}

func (ss *sessionStore) checkSessionLifespan(userID string) (ok bool) {
	session := ss.sessions[userID]
	diff := time.Since(session.createdAt)
	if ok = diff <= ss.lifespan; ok {
		return true
	}
	return false
}

func (ss *sessionStore) searchSession(userID string) *userSession {
	if ss.sessions[userID] != nil {
		return ss.sessions[userID]
	}
	return nil
}

func (bot *bot) createBot() error {
	var err error
	bot.client, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
	if err != nil {
		return err
	}
	return err
}

func (app *app) watchSessions(interval time.Duration) error {
	ss := app.sessionStore
	for range time.Tick(interval) {
		for userID := range ss.sessions {
			if ok := ss.checkSessionLifespan(userID); !ok {
				if err := app.cancelOrder(userID); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
