package firestore

import (
	"time"
)

func parseTime(time time.Time) string {
	return time.Format("20060102")
}
