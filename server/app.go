package main

import (
	"time"

	firestore "github.com/shinyamizuno1008/C_Delivery/server/firestore"
)

type App struct {
	Client    *Client
	Firestore *firestore.Client
	Service   *Service
}

type Service struct {
	Menu          firestore.Menu
	Locations     []firestore.Location
	BusinessHours BusinessHours
	DetailTime    DetailTime
}

type BusinessHours struct {
	Today     string
	Begin     DetailTime `firestore:"begin,omitempty"`
	End       DetailTime `firestore:"end,omitempty"`
	Interval  int        `firestore:"interval,omitempty"`
	Lastorder string     `firestore:"lastorder,omitempty"`
}

type DetailTime struct {
	Hour   int
	Minute int
}

func (app *App) watchSessions(interval time.Duration) error {
	ss := app.Client.SessionStore
	for range time.Tick(interval) {
		for userID := range ss.Sessions {
			if ok := ss.checkSessionLifespan(userID); !ok {
				if err := app.Firestore.CancelOrder(userID); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
