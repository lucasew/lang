package it

import "time"

// NewYearDateFilter ports NewYearDateFilter year-mismatch helpers.
type NewYearDateFilter struct {
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

func (f *NewYearDateFilter) ShouldFlag(year, month int) bool {
	if !f.isJanuary() || month == 12 {
		return false
	}
	return year == f.currentYear()-1
}
