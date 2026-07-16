package rules

import "fmt"

// MatchPosition ports org.languagetool.rules.MatchPosition.
type MatchPosition struct {
	Start int
	End   int
}

func NewMatchPosition(start, end int) MatchPosition {
	return MatchPosition{Start: start, End: end}
}

func (p MatchPosition) GetStart() int { return p.Start }
func (p MatchPosition) GetEnd() int   { return p.End }
func (p MatchPosition) String() string {
	return fmt.Sprintf("%d-%d", p.Start, p.End)
}

func (p MatchPosition) Equals(o MatchPosition) bool {
	return p.Start == o.Start && p.End == o.End
}
