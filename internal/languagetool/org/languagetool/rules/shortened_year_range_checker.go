package rules

import "strconv"

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
		return false
	}
	centuryPrefix := xStr[:2]
	y, err := strconv.Atoi(centuryPrefix + yStr)
	if err != nil {
		return false
	}
	return x >= y
}
