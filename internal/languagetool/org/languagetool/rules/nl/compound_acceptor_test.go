package nl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompoundAcceptor_LoadDefaultWordLists(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadDefaultWordLists())
	require.NotEmpty(t, c.NoS, "no_s.txt")
	require.NotEmpty(t, c.NeedsS, "needs_s.txt")
	require.NotEmpty(t, c.AlwaysNeedsS, "always_needs_s.txt")
	require.NotEmpty(t, c.AlwaysNeedsHyphen, "always_needs_hyphen.txt")
	require.NotEmpty(t, c.Directions, "directions.txt")
	require.NotEmpty(t, c.Part1Exceptions, "part1_exceptions.txt")
	require.NotEmpty(t, c.Part2Exceptions, "part2_exceptions.txt")
	require.NotEmpty(t, c.AcronymExceptions, "acronym_exceptions.txt")
	// case-sensitive directions (Java contains(part1) exact)
	_, ok := c.Directions["Noord-"]
	require.True(t, ok)
	// alwaysNeedsS are suffixes
	_, ok = c.AlwaysNeedsS["ings"]
	require.True(t, ok)
}

func TestCompoundAcceptor_AcceptCompoundParts_NoS(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadNoS(strings.NewReader("straat\npapier\n")))
	// inject tagger/speller like Java DutchTagger + MorfologikDutchSpellerRule
	c.IsNoun = func(w string) bool {
		return w == "puzzel" || w == "versnipperaar"
	}
	c.SpellingOk = func(w string) bool {
		switch w {
		case "straat", "puzzel", "papier", "versnipperaar":
			return true
		default:
			return false
		}
	}
	require.True(t, c.AcceptCompoundParts("straat", "puzzel"))
	require.True(t, c.Accept("straatpuzzel"))
	require.True(t, c.AcceptCompoundParts("papier", "versnipperaar"))
	// colliding vowels rejected on default branch
	c.NoS["politie"] = struct{}{}
	c.IsNoun = func(w string) bool { return w == "eenheid" }
	c.SpellingOk = func(w string) bool { return w == "politie" || w == "eenheid" }
	require.False(t, c.AcceptCompoundParts("politie", "eenheid"))
	// with hyphen + colliding vowels accepted
	require.True(t, c.AcceptCompoundParts("politie", "-eenheid"))
}

func TestCompoundAcceptor_AcceptCompoundParts_NeedsS(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadNeedsS(strings.NewReader("bedrijfs\n")))
	c.IsNoun = func(w string) bool { return w == "regels" || w == "brommer" }
	c.SpellingOk = func(w string) bool {
		switch w {
		case "bedrijf", "regels", "brommer":
			return true
		default:
			return false
		}
	}
	require.True(t, c.AcceptCompoundParts("bedrijfs", "regels"))
	require.True(t, c.Accept("bedrijfsregels"))
	// without connecting s
	require.False(t, c.AcceptCompoundParts("bedrijf", "regels"))
	require.False(t, c.Accept("bedrijfregels"))
}

func TestCompoundAcceptor_AlwaysNeedsS_Suffix(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadAlwaysNeedsS(strings.NewReader("schaps\nings\n")))
	c.IsNoun = func(w string) bool { return w == "blijheid" }
	c.IsExistingWord = func(w string) bool { return w == "zwangerschap" }
	c.SpellingOk = func(w string) bool { return w == "blijheid" }
	// zwangerschaps + blijheid (part1 ends with s, ends with schaps)
	require.True(t, c.AcceptCompoundParts("zwangerschaps", "blijheid"))
}

func TestCompoundAcceptor_AcronymAndHyphen(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadAcronymExceptions(strings.NewReader("wifi\naids\n")))
	require.NoError(t, c.LoadAlwaysNeedsHyphen(strings.NewReader("aspirant-\ncollega-\n")))
	c.SpellingOk = func(w string) bool {
		switch w {
		case "akkoord", "finale", "buschauffeur", "burgemeester":
			return true
		default:
			return false
		}
	}
	// IRA- style uppercase acronym (not in exceptions)
	require.True(t, c.AcceptCompoundParts("IRA-", "akkoord"))
	require.True(t, c.AcceptCompoundParts("WK-", "finale"))
	// WIFI- rejected via acronym exception upper match
	require.False(t, c.AcceptCompoundParts("WIFI-", "verbinding"))
	// alwaysNeedsHyphen
	require.True(t, c.AcceptCompoundParts("aspirant-", "buschauffeur"))
	require.True(t, c.AcceptCompoundParts("collega-", "burgemeester"))
}

func TestCompoundAcceptor_Directions(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadDirections(strings.NewReader("Noord-\nZuidoost-\n")))
	c.IsGeographical = func(w string) bool {
		return w == "Afghanistan" || w == "Turkije"
	}
	require.True(t, c.AcceptCompoundParts("Noord-", "Afghanistan"))
	require.True(t, c.Accept("Noord-Afghanistan"))
	require.False(t, c.AcceptCompoundParts("Noord-", "Frank"))
}

func TestCompoundAcceptor_MaxLengthAndEmpty(t *testing.T) {
	c := NewCompoundAcceptor()
	c.IsNoun = func(string) bool { return true }
	c.SpellingOk = func(string) bool { return true }
	require.False(t, c.Accept(""))
	require.False(t, c.Accept(strings.Repeat("a", 40)))
}

func TestDefaultCompoundAcceptor_Loaded(t *testing.T) {
	require.NotNil(t, DefaultCompoundAcceptor)
	// constructor load may find inspiration resources
	if len(DefaultCompoundAcceptor.NoS) == 0 {
		t.Skip("compound_acceptor resources not discoverable in this environment")
	}
	require.Contains(t, DefaultCompoundAcceptor.NoS, "straat")
}
