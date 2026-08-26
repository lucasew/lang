package rules

import "strconv"

// DateRangeChecker ports org.languagetool.rules.DateRangeChecker.
// Keeps the match when x >= y (invalid ascending range).
type DateRangeChecker struct{}

func NewDateRangeChecker() *DateRangeChecker { return &DateRangeChecker{} }

// Accept returns true when the match should be kept (x >= y).
// Non-numeric values suppress the match (return false).
func (c *DateRangeChecker) Accept(xStr, yStr string) bool {
	x, err1 := strconv.Atoi(xStr)
	y, err2 := strconv.Atoi(yStr)
	if err1 != nil || err2 != nil {
		return false
	}
	return x >= y
}
