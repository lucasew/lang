package rules

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// NewYearDateFilterCore ports AbstractNewYearDateFilter acceptance logic.
type NewYearDateFilterCore struct {
	GetMonth func(localized string) (int, error)
	// ForceJanuary / ForceYear override calendar (tests / Force* on language filters).
	ForceJanuary *bool
	ForceYear    *int
}

func (f *NewYearDateFilterCore) isJanuary() bool {
	if f.ForceJanuary != nil {
		return *f.ForceJanuary
	}
	if IsTest() {
		return true
	}
	return time.Now().Month() == time.January
}

func (f *NewYearDateFilterCore) currentYear() int {
	if f.ForceYear != nil {
		return *f.ForceYear
	}
	if IsTest() {
		return 2014
	}
	return time.Now().Year()
}

// ShouldFlag is true in January for non-December dates whose year is currentYear-1.
func (f *NewYearDateFilterCore) ShouldFlag(year, month int) bool {
	if !f.isJanuary() || month == 12 {
		return false
	}
	return year+1 == f.currentYear()
}

// FormatMessage replaces {year} and {realYear} placeholders.
func (f *NewYearDateFilterCore) FormatMessage(message string, yearFromText int) string {
	msg := strings.ReplaceAll(message, "{year}", strconv.Itoa(yearFromText))
	msg = strings.ReplaceAll(msg, "{realYear}", strconv.Itoa(f.currentYear()))
	return msg
}

// AcceptFromArgs returns rewritten message when the new-year condition holds; "" suppresses.
func (f *NewYearDateFilterCore) AcceptFromArgs(args map[string]string, message string) string {
	y, err := strconv.Atoi(args["year"])
	if err != nil {
		return ""
	}
	m, err := f.parseMonth(args["month"])
	if err != nil {
		return ""
	}
	if !f.ShouldFlag(y, m) {
		return ""
	}
	return f.FormatMessage(message, y)
}

func (f *NewYearDateFilterCore) parseMonth(monthStr string) (int, error) {
	monthStr = strings.TrimSpace(monthStr)
	allDigit := true
	for _, r := range monthStr {
		if !unicode.IsDigit(r) {
			allDigit = false
			break
		}
	}
	if allDigit {
		return strconv.Atoi(monthStr)
	}
	if f.GetMonth != nil {
		return f.GetMonth(monthStr)
	}
	return 0, fmt.Errorf("non-numeric month without GetMonth")
}
