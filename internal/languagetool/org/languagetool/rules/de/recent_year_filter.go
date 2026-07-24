package de

import (
	"strconv"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

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

// AcceptRuleMatch ports RecentYearFilter.acceptRuleMatch.
func (f *RecentYearFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	year, err1 := strconv.Atoi(arguments["year"])
	maxBack, err2 := strconv.Atoi(arguments["maxYearsBack"])
	if err1 != nil || err2 != nil {
		return nil
	}
	if f.Accept(year, maxBack) {
		return match
	}
	return nil
}
