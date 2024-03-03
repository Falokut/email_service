package utils

import (
	"math"
	"strconv"
)

const (
	second          = 1
	secondsInMinute = 60 * second
	secondsInHour   = 60 * secondsInMinute
	secondsInDay    = 24 * secondsInHour
	secondsInWeek   = 7 * secondsInDay
)

func ResolveTime(timeInSeconds float64) string {
	if timeInSeconds == second {
		return "1 секунда"
	}

	if timeInSeconds <= 1.5*secondsInMinute {
		return strconv.Itoa(int(timeInSeconds)) + " секунд"
	}

	if timeInSeconds <= 1.5*secondsInHour {
		minutes := math.Round(timeInSeconds / secondsInMinute)
		return strconv.Itoa(int(minutes)) + " минут"
	}

	if timeInSeconds < 24*secondsInHour {
		hours := math.Round(timeInSeconds / secondsInHour)
		return strconv.Itoa(int(hours)) + " часов"
	}

	if timeInSeconds == secondsInDay {
		return "1 день"
	}

	if timeInSeconds < secondsInWeek {
		days := math.Round(timeInSeconds / secondsInDay)
		return strconv.Itoa(int(days)) + " дней"
	}

	weeks := math.Round(timeInSeconds / secondsInWeek)
	return strconv.Itoa(int(weeks)) + " недель"
}
