package main

import (
	"fmt"
	"time"

	"github.com/shinyamizuno1008/C_Delivery/server/firestore"
)

type SessionStore struct {
	Sessions Sessions
	Lifespan time.Duration
}

type Sessions map[string]*UserSession

type UserSession struct {
	OrderID   string
	PrevStep  int
	CreatedAt time.Time
	Products  firestore.Products
}

func (ss *SessionStore) createSession(userID string) *UserSession {
	ss.Sessions[userID] = &UserSession{PrevStep: Begin, CreatedAt: time.Now(), Products: make(firestore.Products)}
	return ss.Sessions[userID]
}

func (ss *SessionStore) deleteUserSession(userID string) error {
	if ss.Sessions[userID] == nil {
		return fmt.Errorf("User doesn't exist in session Store: ID. %v", userID)
	}
	delete(ss.Sessions, userID)
	return nil
}

func (ss *SessionStore) checkSessionLifespan(userID string) (ok bool) {
	session := ss.Sessions[userID]
	diff := time.Since(session.CreatedAt)
	if ok = diff <= ss.Lifespan; ok {
		return true
	}
	return false
}

func (ss *SessionStore) searchSession(userID string) *UserSession {
	if ss.Sessions[userID] != nil {
		return ss.Sessions[userID]
	}
	return nil
}
