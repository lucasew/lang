package implementation

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestItem_CompareTo(t *testing.T) {
	// Lower distance first
	lowDist := NewSuggestItem("a", 1, 1)
	highDist := NewSuggestItem("b", 2, 100)
	require.Equal(t, -1, lowDist.CompareTo(highDist))
	require.Equal(t, 1, highDist.CompareTo(lowDist))

	// Same distance: higher count first (Long.compare(other.count, this.count))
	moreFreq := NewSuggestItem("x", 1, 10)
	lessFreq := NewSuggestItem("y", 1, 5)
	require.Equal(t, -1, moreFreq.CompareTo(lessFreq))
	require.Equal(t, 1, lessFreq.CompareTo(moreFreq))

	// Equal distance and count → 0 (even if terms differ — Java does not compare term)
	a := NewSuggestItem("foo", 1, 7)
	b := NewSuggestItem("bar", 1, 7)
	require.Equal(t, 0, a.CompareTo(b))
	require.Equal(t, 0, b.CompareTo(a))
	require.Equal(t, 0, a.CompareTo(a))

	// Count edge: zero and negative (Long.compare)
	zero := NewSuggestItem("z", 0, 0)
	neg := NewSuggestItem("n", 0, -1)
	require.Equal(t, -1, zero.CompareTo(neg)) // 0 > -1 → other.count smaller when comparing zero vs neg from zero? Long.compare(neg, zero) = Long.compare(-1, 0) = -1, so zero.CompareTo(neg) = -1 → zero sorts before neg (higher count first)
	require.Equal(t, 1, neg.CompareTo(zero))
}

func TestSuggestItem_Compare_DelegatesToCompareTo(t *testing.T) {
	a := NewSuggestItem("a", 1, 10)
	b := NewSuggestItem("b", 1, 5)
	c := NewSuggestItem("c", 2, 100)
	var cmp SuggestItem
	require.Equal(t, a.CompareTo(b), cmp.Compare(a, b))
	require.Equal(t, a.CompareTo(c), cmp.Compare(a, c))
	require.Equal(t, b.CompareTo(a), cmp.Compare(b, a))
}

func TestSuggestItem_Less_Wrapper(t *testing.T) {
	a := NewSuggestItem("a", 1, 10)
	b := NewSuggestItem("b", 1, 5)
	c := NewSuggestItem("c", 2, 100)
	require.True(t, a.Less(b))
	require.True(t, a.Less(c))
	require.False(t, b.Less(a))
	require.False(t, a.Less(a))
}

func TestSuggestItem_SortStabilityMatrix(t *testing.T) {
	// Mix of distances and counts; expected order by CompareTo (distance asc, count desc)
	items := []SuggestItem{
		NewSuggestItem("d2-low", 2, 1),
		NewSuggestItem("d1-high", 1, 100),
		NewSuggestItem("d0", 0, 3),
		NewSuggestItem("d1-mid", 1, 50),
		NewSuggestItem("d1-low", 1, 1),
		NewSuggestItem("d2-high", 2, 99),
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Less(items[j])
	})
	wantTerms := []string{"d0", "d1-high", "d1-mid", "d1-low", "d2-high", "d2-low"}
	got := make([]string, len(items))
	for i, it := range items {
		got[i] = it.Term
	}
	require.Equal(t, wantTerms, got)

	// Equal keys stay in relative order under stable sort
	eq := []SuggestItem{
		NewSuggestItem("first", 1, 5),
		NewSuggestItem("second", 1, 5),
		NewSuggestItem("third", 1, 5),
	}
	sort.SliceStable(eq, func(i, j int) bool { return eq[i].Less(eq[j]) })
	require.Equal(t, "first", eq[0].Term)
	require.Equal(t, "second", eq[1].Term)
	require.Equal(t, "third", eq[2].Term)
}

func TestSuggestItem_Equals_TermOnly(t *testing.T) {
	// Same term, different distance/count → still equal (Java equals uses term only)
	a := NewSuggestItem("word", 1, 10)
	b := NewSuggestItem("word", 9, 999)
	require.True(t, a.Equals(b))
	require.True(t, b.Equals(a))
	require.True(t, a.Equals(a))

	// Different term → not equal even if distance/count match
	c := NewSuggestItem("other", 1, 10)
	require.False(t, a.Equals(c))
	require.False(t, c.Equals(a))

	// Empty terms equal
	require.True(t, NewSuggestItem("", 0, 0).Equals(NewSuggestItem("", 3, 4)))
}

func TestSuggestItem_HashCode_JavaStringHash(t *testing.T) {
	// Java "hello".hashCode() == 99162322
	require.Equal(t, 99162322, NewSuggestItem("hello", 0, 0).HashCode())
	// Empty string hashCode is 0
	require.Equal(t, 0, NewSuggestItem("", 1, 2).HashCode())
	// Same term → same hash regardless of distance/count
	a := NewSuggestItem("term", 1, 1)
	b := NewSuggestItem("term", 5, 99)
	require.Equal(t, a.HashCode(), b.HashCode())
	// Different term → different hash (typical; not a contract but sanity)
	require.NotEqual(t, NewSuggestItem("a", 0, 0).HashCode(), NewSuggestItem("b", 0, 0).HashCode())
	// UTF-16 / non-ASCII: Java "café".hashCode()
	// c=99, a=97, f=102, é=233
	// h=0; 31*0+99=99; 31*99+97=3166; 31*3166+102=98248; 31*98248+233=3045921
	require.Equal(t, 3045921, NewSuggestItem("café", 0, 0).HashCode())
	// Surrogate pair: 😀 is U+1F600 → high 0xD83D, low 0xDE00
	// h = 31*0xD83D + 0xDE00
	emojiHash := int(int32(31*0xD83D + 0xDE00))
	require.Equal(t, emojiHash, NewSuggestItem("😀", 0, 0).HashCode())
	// Equals-consistent: equal terms share hash
	require.Equal(t, a.HashCode(), b.HashCode())
	require.True(t, a.Equals(b))
}

func TestSuggestItem_String_Format(t *testing.T) {
	// "{" + term + ", " + distance + ", " + count + "}"
	require.Equal(t, "{hello, 1, 42}", NewSuggestItem("hello", 1, 42).String())
	require.Equal(t, "{, 0, 0}", NewSuggestItem("", 0, 0).String())
	require.Equal(t, "{-1, 2, -3}", NewSuggestItem("-1", 2, -3).String())
}

func TestSuggestItem_Clone_Independence(t *testing.T) {
	orig := NewSuggestItem("orig", 2, 100)
	cl := orig.Clone()
	require.Equal(t, orig.Term, cl.Term)
	require.Equal(t, orig.Distance, cl.Distance)
	require.Equal(t, orig.Count, cl.Count)
	require.True(t, orig.Equals(cl))
	require.Equal(t, 0, orig.CompareTo(cl))

	// Mutating clone fields must not affect original (value type; reassign clone)
	cl.Term = "changed"
	cl.Distance = 99
	cl.Count = -1
	require.Equal(t, "orig", orig.Term)
	require.Equal(t, 2, orig.Distance)
	require.Equal(t, int64(100), orig.Count)
	require.False(t, orig.Equals(cl))
}
