package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckPostagsInSuggestionFilter(t *testing.T) {
	f := NewCheckPostagsInSuggestionFilter(func(tok string) []string {
		switch tok {
		case "the":
			return []string{"DT"}
		case "cat":
			return []string{"NN"}
		case "runs":
			return []string{"VBZ"}
		default:
			return []string{"XX"}
		}
	})
	got := f.Filter([]string{"the cat", "runs cat", "the runs"}, "DT,NN")
	require.Equal(t, []string{"the cat"}, got)
}

func TestCheckPostagsInSuggestionFilter_FullPosMatch(t *testing.T) {
	// Java matchesPosTagRegex uses Matcher.matches (full); invent partial MatchString would accept "DTX".
	f := NewCheckPostagsInSuggestionFilter(func(tok string) []string {
		return []string{"DTX"}
	})
	require.Empty(t, f.Filter([]string{"the"}, "DT"))
	require.Equal(t, []string{"the"}, f.Filter([]string{"the"}, "DT.*"))
}

func TestCheckPostagsInSuggestionFilter_SplitWSLikeJava(t *testing.T) {
	// Java split("\\s+") keeps leading empty; Fields invent would not.
	f := NewCheckPostagsInSuggestionFilter(func(tok string) []string {
		if tok == "" {
			return []string{"EMPTY"}
		}
		if tok == "cat" {
			return []string{"NN"}
		}
		return []string{"XX"}
	})
	// " cat" → ["", "cat"] with Java split
	require.Equal(t, []string{" cat"}, f.Filter([]string{" cat"}, "EMPTY,NN"))
}

// Java Pattern \\s without UNICODE_CHARACTER_CLASS does not split on NBSP.
func TestCheckPostagsInSuggestionFilter_NBSPNotSplit(t *testing.T) {
	f := NewCheckPostagsInSuggestionFilter(func(tok string) []string {
		if tok == "the\u00a0cat" {
			return []string{"NN"}
		}
		return []string{"XX"}
	})
	// One token (NBSP interior) vs one tag — keep; Unicode \\s would invent two tokens → panic.
	require.Equal(t, []string{"the\u00a0cat"}, f.Filter([]string{"the\u00a0cat"}, "NN"))
}

func TestCheckPostagsInSuggestionFilter_MismatchPanics(t *testing.T) {
	f := NewCheckPostagsInSuggestionFilter(func(tok string) []string { return []string{"NN"} })
	// Java IOException on token/tag count mismatch — not invent skip
	// one token vs two tags
	require.Panics(t, func() {
		f.Filter([]string{"only"}, "DT,NN")
	})
}

func TestCheckPostagsInSuggestionFilter_NoTaggerPanics(t *testing.T) {
	// Java throws if tagger null — not invent soft drop
	f := NewCheckPostagsInSuggestionFilter(nil)
	require.Panics(t, func() {
		f.Filter([]string{"x"}, "N.*")
	})
}
