package main

import (
	"net/http"
	"time"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

type bot struct {
	request *http.Request
	writer  http.ResponseWriter
}

type app struct {
	bot     *linebot.Client
	client       *firebase.App
	sessionStore sessionStore
}

type sessionStore struct {
	sessions map[string]userInfo
	lifespan time.Duration
}

type userInfo struct {
	prevStep  string
	createdAt time.Time
	order     Order
}

func (app *app) createOrderSession() (userID string) {
	app.sessionStore.sessions[userID] = userInfo{createdAt: time.Now()}
	return
}

func (app *app) checkSessionLifespan(userID string) (ok bool) {
	diff := time.Since(app.sessionStore.sessions[userID].createdAt)
	if ok = diff <= app.sessionStore.lifespan; ok {
		return true
	}
	delete(app.sessionStore.sessions, userID)
	return false
}

func (app *app) SessionHandler(f gin.HandlerFunc) gin.HandlerFunc {
	sessionStore := app.sessionStore
	return func(g *gin.Context) {
		if !sessionStore.(g.Query(SESSION)) {
			log.Errorf("Session not found: %s", g.Request.URL)
			ms.errorResult(g, http.StatusInternalServerError, errorMessage(SESSION))
			return
		}
		f(g)
		return
	}
}
