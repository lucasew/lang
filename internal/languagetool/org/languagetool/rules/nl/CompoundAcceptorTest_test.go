package nl

// Twin of CompoundAcceptorTest — hooks inject noun/speller (full Dutch tagger/dict optional).
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of CompoundAcceptorTest.testAcceptCompound (shape; resources + hooks).
func TestCompoundAcceptor_AcceptCompound(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadDefaultWordLists())
	if len(c.NoS) == 0 {
		t.Skip("compound_acceptor lists not found")
	}
	// inject POS + spelling like a wired DutchTagger + speller
	nouns := map[string]struct{}{
		"puzzel": {}, "versnipperaar": {}, "regels": {}, "brommer": {},
		"schip": {}, "agente": {},
	}
	c.IsNoun = func(w string) bool {
		_, ok := nouns[strings.ToLower(w)]
		if !ok {
			_, ok = nouns[w]
		}
		return ok
	}
	c.IsExistingWord = func(w string) bool { return true }
	c.SpellingOk = func(w string) bool {
		// accept normal-case stems used in Java test names
		return normalCasePattern.MatchString(w) || normalCasePattern.MatchString(strings.ToLower(w))
	}
	// Java: straatpuzzel (no_s + noun)
	if _, ok := c.NoS["straat"]; ok {
		require.True(t, c.Accept("straatpuzzel"), "straatpuzzel")
	}
	// Java: bedrijfsregels (needs_s)
	if _, ok := c.NeedsS["bedrijfs"]; ok {
		require.True(t, c.Accept("bedrijfsregels"), "bedrijfsregels")
		require.False(t, c.Accept("bedrijfregels"), "bedrijfregels")
	}
	// too long
	require.False(t, c.Accept(strings.Repeat("x", 40)))
	require.False(t, c.Accept(""))
}

// Port of CompoundAcceptorTest.testAcceptCompoundInternal
func TestCompoundAcceptor_AcceptCompoundInternal(t *testing.T) {
	c := NewCompoundAcceptor()
	c.IsNoun = func(w string) bool { return w == "schip" }
	// Java spellingOk(part1 without trailing s) for needs_s branch → "passagier"
	c.SpellingOk = func(w string) bool {
		switch w {
		case "passagier", "passagiers", "schip", "papier", "versnipperaar":
			return true
		default:
			return false
		}
	}
	require.NoError(t, c.LoadNeedsS(strings.NewReader("passagiers\n")))
	require.True(t, c.AcceptCompoundParts("passagiers", "schip"))
	require.NoError(t, c.LoadNoS(strings.NewReader("papier\n")))
	c.IsNoun = func(w string) bool { return w == "versnipperaar" }
	require.True(t, c.AcceptCompoundParts("papier", "versnipperaar"))
	// colliding vowels
	c.NoS["politie"] = struct{}{}
	c.IsNoun = func(w string) bool { return w == "eenheid" }
	c.SpellingOk = func(w string) bool { return true }
	require.False(t, c.AcceptCompoundParts("politie", "eenheid"))
	require.True(t, c.AcceptCompoundParts("politie", "-eenheid"))
	// length gate
	require.False(t, c.Accept(strings.Repeat("a", 40)))
}
