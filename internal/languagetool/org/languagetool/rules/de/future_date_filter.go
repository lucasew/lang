package de

import (
	"strconv"
	"time"
)

// FutureDateFilter ports FutureDateFilter date assembly for pattern filters.
// Full RuleFilter.accept integration is deferred.
type FutureDateFilter struct {
	helper *DateFilterHelper
}

func NewFutureDateFilter() *FutureDateFilter {
	return &FutureDateFilter{helper: NewDateFilterHelper()}
}

// IsFuture reports whether year/month/day (1-based month) is strictly after now (UTC date).
func (f *FutureDateFilter) IsFuture(year, month, day int) bool {
	if year <= 0 || month < 1 || month > 12 || day < 1 {
		return false
	}
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return d.After(today)
}

// ParseDayOfMonth extracts leading digits from strings like "23." / "23".
func ParseDayOfMonth(s string) (int, error) {
	n := ""
	for _, r := range s {
		if r >= '0' && r <= '9' {
			n += string(r)
		} else if n != "" {
			break
		}
	}
	return strconv.Atoi(n)
}
