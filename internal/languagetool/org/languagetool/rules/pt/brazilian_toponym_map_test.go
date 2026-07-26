package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBrazilianToponymMap_NewAndLoad(t *testing.T) {
	m1 := NewBrazilianToponymMap()
	require.NotNil(t, m1)
	require.NotNil(t, m1.map_)
	require.NotEmpty(t, m1.map_["SP"])
	require.NotEmpty(t, m1.all)

	// Load is a process singleton; New is a fresh map each call.
	m2 := LoadBrazilianToponymMap()
	m3 := LoadBrazilianToponymMap()
	require.Same(t, m2, m3)
	m4 := NewBrazilianToponymMap()
	require.NotSame(t, m1, m4)
}

func TestBrazilianToponymMap_IsValidToponym(t *testing.T) {
	m := NewBrazilianToponymMap()

	// Real municipalities (case/hyphen via toponymIter normalise path).
	require.True(t, m.IsValidToponym("São Paulo"))
	require.True(t, m.IsValidToponym("são paulo"))
	require.True(t, m.IsValidToponym("Niterói"))
	require.True(t, m.IsValidToponym("Rio de Janeiro"))

	// Hungry left-strip: prefix + real suffix municipality.
	require.True(t, m.IsValidToponym("Venho do Rio de Janeiro"))
	require.True(t, m.IsValidToponym("prefix São Paulo"))
	require.True(t, m.IsValidToponym("do-Rio-de-Janeiro")) // hyphen→space then suffix

	// Garbage.
	require.False(t, m.IsValidToponym("Narnia"))
	require.False(t, m.IsValidToponym("Oogabooga"))
	require.False(t, m.IsValidToponym(""))
}

func TestBrazilianToponymMap_ToponymIterHungryLeftStrip(t *testing.T) {
	m := NewBrazilianToponymMap()

	// Walk left-strip order: full phrase first, then drop leftmost tokens.
	var seen []string
	_ = m.toponymIter("Alpha Beta Gamma", func(check string) bool {
		seen = append(seen, check)
		return false // never accept — exercise full walk
	})
	require.Equal(t, []string{
		"alpha beta gamma",
		"beta gamma",
		"gamma",
	}, seen)

	// Double spaces: Java split(" ") keeps middle empties.
	seen = nil
	_ = m.toponymIter("a  b", func(check string) bool {
		seen = append(seen, check)
		return false
	})
	require.Equal(t, []string{
		"a  b", // join ["a","","b"]
		" b",   // join ["","b"]
		"b",
	}, seen)

	// Accept mid-walk (first true wins) — "rio de janeiro" is real.
	// "Venho do Rio de Janeiro" → checks until "rio de janeiro".
	var hit string
	ok := m.toponymIter("Venho do Rio de Janeiro", func(check string) bool {
		if _, found := m.all[check]; found {
			hit = check
			return true
		}
		return false
	})
	require.True(t, ok)
	require.Equal(t, "rio de janeiro", hit)
}

func TestBrazilianToponymMap_StatesWithMunicipality_AsIs(t *testing.T) {
	m := NewBrazilianToponymMap()

	// Stored form is loader-normalised (lower, hyphen→space). Pass AS-IS as Java does.
	states := m.StatesWithMunicipality("são paulo")
	require.Contains(t, states, "SP")
	// No invent: raw surface form that is not stored does not match.
	require.Empty(t, m.StatesWithMunicipality("São Paulo"))
	require.Empty(t, m.StatesWithMunicipality("Sao Paulo"))
	require.Empty(t, m.StatesWithMunicipality("são-paulo")) // stored as space, not hyphen

	// Niterói is RJ only (normalised).
	require.Equal(t, []string{"RJ"}, sortedCopy(m.StatesWithMunicipality("niterói")))
	require.Empty(t, m.StatesWithMunicipality("Niterói"))

	// Unknown.
	require.Empty(t, m.StatesWithMunicipality("narnia"))
}

func TestBrazilianToponymMap_IsToponymInState_AsIs(t *testing.T) {
	m := NewBrazilianToponymMap()

	// Exact contains semantics — already-normalised form.
	require.True(t, m.IsToponymInState("são paulo", "SP"))
	require.True(t, m.IsToponymInState("niterói", "RJ"))
	require.False(t, m.IsToponymInState("niterói", "SP"))

	// Unnormalised does NOT invent-match.
	require.False(t, m.IsToponymInState("São Paulo", "SP"))
	require.False(t, m.IsToponymInState("Niterói", "RJ"))
	require.False(t, m.IsToponymInState("são-paulo", "SP"))

	// Missing state key → false (Java: municipalities != null && …).
	require.False(t, m.IsToponymInState("são paulo", "XX"))
	require.False(t, m.IsToponymInState("são paulo", ""))
}

func TestBrazilianToponymMap_JavaStringSplitSpace(t *testing.T) {
	// Twin of Java String.split(" ") limit 0.
	require.Equal(t, []string{""}, javaStringSplitSpace(""))
	require.Equal(t, []string{}, javaStringSplitSpace(" "))
	require.Equal(t, []string{}, javaStringSplitSpace("  "))
	require.Equal(t, []string{"a"}, javaStringSplitSpace("a"))
	require.Equal(t, []string{"a"}, javaStringSplitSpace("a "))
	require.Equal(t, []string{"", "a"}, javaStringSplitSpace(" a"))
	require.Equal(t, []string{"a", "b"}, javaStringSplitSpace("a b"))
	require.Equal(t, []string{"a", "", "b"}, javaStringSplitSpace("a  b"))
	require.Equal(t, []string{"a", "b"}, javaStringSplitSpace("a b "))
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	// tiny insertion sort to avoid extra import noise for 0–1 elements typical here
	for i := 1; i < len(out); i++ {
		j := i
		for j > 0 && out[j-1] > out[j] {
			out[j-1], out[j] = out[j], out[j-1]
			j--
		}
	}
	return out
}
