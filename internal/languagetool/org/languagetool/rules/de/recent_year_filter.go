package de

import "time"

// RecentYearFilter ports org.languagetool.rules.de.RecentYearFilter.
// Keeps matches when year is in [currentYear-maxYearsBack, currentYear).
type RecentYearFilter struct {
	// ForceYear overrides current calendar year for tests.
	ForceYear *int
}

func NewRecentYearFilter() *RecentYearFilter {
	return &RecentYearFilter{}
}

func (f *RecentYearFilter) Accept(year, maxYearsBack int) bool {
	thisYear := time.Now().Year()
	if f.ForceYear != nil {
		thisYear = *f.ForceYear
	}
	maxYear := thisYear - maxYearsBack
	return year < thisYear && year >= maxYear
}
