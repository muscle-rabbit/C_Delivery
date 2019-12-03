package main

import (
	"regexp"
)

func validateBegin(text string) bool {
	return text == "予約開始"
}

func validateDate(text string) (bool, error) {
	// 例: 本日 11/14 (木)
	matched, err := regexp.Match(`.. (0[1-9]|1[0-2])/(0[1-9]|[12]\d|3[01]) \(.\)`, []byte(text))
	return matched, err
}

func validateTime(text string) (bool, error) {
	// 例：12:00AM~12:59AM
	matched, err := regexp.Match(`([01]?\d|2[0-3]):([0-5]?\d)?[0-5][0-9][AP]M~([01]?\d|2[0-3]):([0-5]?\d)?[0-5][0-9][AP]M`, []byte(text))
	return matched, err
}

func validateProduct(text string) (bool, error) {
	// 例: 唐揚げ弁当x1
	matched, err := regexp.Match(`(^.*x[0-9]{1,2}|注文決定)`, []byte(text))
	return matched, err
}

func validateLocation(text string, locations []Location) bool {
	// 例: 6号館 1F
	for _, location := range locations {
		if text == location.Name {
			return true
		}
	}
	return false
}

func validateConfirmation(text string) bool {
	return text == "はい" || text == "いいえ"
}
