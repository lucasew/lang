package de

import "time"

// NewYearDateFilter ports org.languagetool.rules.de.NewYearDateFilter helpers
// for pattern-rule year mismatch detection in early January.
type NewYearDateFilter struct {
	// ForceJanuary / ForceYear override calendar for tests (Java TestHackHelper).
	ForceJanuary *bool
	ForceYear    *int
}

func NewNewYearDateFilter() *NewYearDateFilter {
	return &NewYearDateFilter{}
}

func (f *NewYearDateFilter) isJanuary() bool {
	if f.ForceJanuary != nil {
		return *f.ForceJanuary
	}
	return time.Now().Month() == time.January
}

func (f *NewYearDateFilter) currentYear() int {
	if f.ForceYear != nil {
		return *f.ForceYear
	}
	return time.Now().Year()
}

// ShouldFlag reports whether a date with year/month may still refer to the previous
// year wrongly (Java: January + year == currentYear-1 + month != December).
func (f *NewYearDateFilter) ShouldFlag(year, month int) bool {
	if !f.isJanuary() {
		return false
	}
	if month == 12 {
		return false
	}
	return year == f.currentYear()-1
}

// MonthNumber uses DateFilterHelper (1–12).
func (f *NewYearDateFilter) MonthNumber(localizedMonth string) (int, error) {
	m, err := NewDateFilterHelper().GetMonth(localizedMonth)
	return int(m), err
}
