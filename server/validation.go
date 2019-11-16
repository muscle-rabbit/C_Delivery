package main

import (
	"regexp"
)

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
