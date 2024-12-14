package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseDuration parses human-readable time format into time.Duration
func parseDuration(input string) (time.Duration, error) {
	re := regexp.MustCompile(`(?i)(\d+\s*нед|\d+\s*день|\d+\s*дня|\d+\s*дней|\d+\s*час|\d+\s*мин|\d+\s*сек)`)
	matches := re.FindAllString(strings.ToLower(input), -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("неверный формат времени")
	}

	var duration time.Duration
	for _, match := range matches {
		var value int
		var err error

		if strings.Contains(match, "день") || strings.Contains(match, "дня") || strings.Contains(match, "дней") {
			value, err = extractNumber(match, "день")
			duration += time.Duration(value) * 24 * time.Hour
		} else if strings.Contains(match, "час") {
			value, err = extractNumber(match, "час")
			duration += time.Duration(value) * time.Hour
		} else if strings.Contains(match, "мин") {
			value, err = extractNumber(match, "мин")
			duration += time.Duration(value) * time.Minute
		} else if strings.Contains(match, "сек") {
			value, err = extractNumber(match, "сек")
			duration += time.Duration(value) * time.Second
		} else if strings.Contains(match, "нед") {
			value, err = extractNumber(match, "нед")
			duration += time.Duration(value) * 24 * 7 * time.Hour
		}

		if err != nil {
			return 0, err
		}
	}

	return duration, nil
}

// extractNumber extracts the number from the time string
func extractNumber(input, unit string) (int, error) {
	value := strings.TrimSpace(strings.Replace(input, unit, "", 1))
	return strconv.Atoi(value)
}
