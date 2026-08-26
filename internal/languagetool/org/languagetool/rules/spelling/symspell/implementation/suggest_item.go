package implementation

import "fmt"

// SuggestItem ports org.languagetool.rules.spelling.symspell.implementation.SuggestItem.
// Java: public class SuggestItem implements Comparator<SuggestItem>, Comparable<SuggestItem>
type SuggestItem struct {
	// Term is the suggested correctly spelled word (Java: public String term).
	Term string
	// Distance is edit distance between searched-for word and suggestion (Java: public int distance).
	Distance int
	// Count is frequency of suggestion in the dictionary (Java: public long count).
	Count int64
}

// NewSuggestItem ports SuggestItem(String term, int distance, long count).
func NewSuggestItem(term string, distance int, count int64) SuggestItem {
	return SuggestItem{Term: term, Distance: distance, Count: count}
}

// Compare ports SuggestItem.compare(SuggestItem, SuggestItem) — delegates to compareTo.
// Java ignores the receiver as Comparator instance state; behavior matches suggestItem.compareTo(t1).
func (SuggestItem) Compare(a, b SuggestItem) int {
	return a.CompareTo(b)
}

// Equals ports SuggestItem.equals(Object): true iff other is SuggestItem and term equals (term only).
func (a SuggestItem) Equals(other SuggestItem) bool {
	return a.Term == other.Term
}

// HashCode ports SuggestItem.hashCode(): term.hashCode() (Java String.hashCode on UTF-16 units).
func (a SuggestItem) HashCode() int {
	return javaStringHashCode(a.Term)
}

// String ports SuggestItem.toString(): "{" + term + ", " + distance + ", " + count + "}".
func (a SuggestItem) String() string {
	return fmt.Sprintf("{%s, %d, %d}", a.Term, a.Distance, a.Count)
}

// CompareTo ports SuggestItem.compareTo(SuggestItem other):
// order by distance ascending, then by frequency count descending.
// If distance equal: Long.compare(other.count, this.count); else Integer.compare(this.distance, other.distance).
func (a SuggestItem) CompareTo(other SuggestItem) int {
	if a.Distance == other.Distance {
		return longCompare(other.Count, a.Count)
	}
	return integerCompare(a.Distance, other.Distance)
}

// Less is a thin wrapper for sort.Slice: true when a should sort before b (CompareTo < 0).
func (a SuggestItem) Less(b SuggestItem) bool {
	return a.CompareTo(b) < 0
}

// Clone ports SuggestItem.clone(): new SuggestItem(term, distance, count).
func (a SuggestItem) Clone() SuggestItem {
	return NewSuggestItem(a.Term, a.Distance, a.Count)
}

// integerCompare ports Java Integer.compare(x, y).
func integerCompare(x, y int) int {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}

// longCompare ports Java Long.compare(x, y).
func longCompare(x, y int64) int {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}
