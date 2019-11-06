package main

import (
	"time"

	firebase "firebase.google.com/go"
	"github.com/line/line-bot-sdk-go/linebot"
)

type app struct {
	bot          *linebot.Client
	client       *firebase.App
	sessionStore *sessionStore
	service      *service
}

type sessionStore struct {
	sessions sessions
	lifespan time.Duration
}

type userSession struct {
	prevStep  int
	createdAt time.Time
	order     Order
}

type sessions map[string]*userSession

type service struct {
	menu          Menu
	locations     []Location
	businessHours businessHours
	detailTime    detailTime
	stockTable    stockTable
}

type Location struct {
	Name string `firestore:"name,omitempty"`
}

type businessHours struct {
	begin     detailTime `firestore:"begin,omitempty"`
	end       detailTime `firestore:"end,omitempty"`
	interval  int        `firestore:"interval,omitempty"`
	lastorder string     `firestore:"lastorder,omitempty"`
}

type detailTime struct {
	hour   int
	minute int
}

type stockTable struct {
	read  chan products
	write chan writeProductsCh
	stock products
}

type products map[string]int

type writeProductsCh struct {
	action   string
	products products
}

func (st *stockTable) run() {
	for {
		select {
		case ch := <-st.write:
			st.stock.push(products)
		case products := <-st.read:
			// Todo: 在庫がなかったら subtract の channel を閉じる。
			// Todo: 在庫がまた復活したら sbutract の channel を開ける。
			st.stock.subtract(products)
		}
	}
}

func (p products) push(new products) {
	for newItem, newN := range new {
		p[newItem] += newN
	}
}

func (p products) subtract(new products) {
	for newItem, newN := range new {
		p[newItem] -= newN
	}
}

func (p products) write(writeProductsCh) {
	action := writeProductsCh.action
	switch writeProductsCh.action {
	case "push":
		p.push(writeProductsCh.products)

	}
}

func (ss *sessionStore) createSession(userID string) *userSession {
	ss.sessions[userID] = &userSession{prevStep: begin, createdAt: time.Now(), order: Order{products: make(products)}}
	return ss.sessions[userID]
}

func (ss *sessionStore) deleteUserSession(userID string) {
	delete(ss.sessions, userID)
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
