package implementation

import "fmt"

// SuggestItem ports org.languagetool.rules.spelling.symspell.implementation.SuggestItem.
type SuggestItem struct {
	Term     string
	Distance int
	Count    int64
}

func NewSuggestItem(term string, distance int, count int64) SuggestItem {
	return SuggestItem{Term: term, Distance: distance, Count: count}
}

// Less reports whether a should sort before b (distance asc, count desc).
func (a SuggestItem) Less(b SuggestItem) bool {
	if a.Distance == b.Distance {
		return a.Count > b.Count
	}
	return a.Distance < b.Distance
}

func (a SuggestItem) EqualTerm(b SuggestItem) bool { return a.Term == b.Term }

func (a SuggestItem) Clone() SuggestItem { return a }

func (a SuggestItem) String() string {
	return fmt.Sprintf("{%s, %d, %d}", a.Term, a.Distance, a.Count)
}
