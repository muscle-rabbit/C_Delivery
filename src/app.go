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

func (ss *sessionStore) createSession(userID string) *userSession {
	ss.sessions[userID] = &userSession{prevStep: begin, createdAt: time.Now()}
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
