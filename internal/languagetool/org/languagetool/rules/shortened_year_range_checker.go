package rules

import (
	"strconv"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ShortenedYearRangeChecker ports org.languagetool.rules.ShortenedYearRangeChecker.
// Interprets y as a two-digit year under x's century prefix (e.g. 1998–92 → 1992).
type ShortenedYearRangeChecker struct{}

func NewShortenedYearRangeChecker() *ShortenedYearRangeChecker {
	return &ShortenedYearRangeChecker{}
}

// Accept returns true when the match should be kept (x >= expanded y).
func (c *ShortenedYearRangeChecker) Accept(xStr, yStr string) bool {
	x, err := strconv.Atoi(xStr)
	if err != nil || len(xStr) < 2 {
		// Java: NumberFormatException / short string → ignore (null match).
		return false
	}
	centuryPrefix := xStr[:2]
	y, err := strconv.Atoi(centuryPrefix + yStr)
	if err != nil {
		return false
	}
	return x >= y
}

// AcceptRuleMatch ports ShortenedYearRangeChecker.acceptRuleMatch.
// Args: x (full year), y (two-digit end year).
func (c *ShortenedYearRangeChecker) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if c == nil || match == nil {
		return nil
	}
	if c.Accept(arguments["x"], arguments["y"]) {
		return match
	}
	return nil
}
