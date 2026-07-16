package en

import "time"

// NewYearDateFilter ports EN NewYearDateFilter year-mismatch helpers.
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
	if !f.isJanuary() {
		return false
	}
	if month == 12 {
		return false
	}
	return year == f.currentYear()-1
}

func (f *NewYearDateFilter) MonthNumber(localizedMonth string) (int, error) {
	m, err := NewDateFilterHelper().GetMonth(localizedMonth)
	return int(m), err
}
