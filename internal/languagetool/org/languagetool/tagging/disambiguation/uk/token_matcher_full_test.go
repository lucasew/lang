package uk

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestTokenMatcher_fullPOSMatch(t *testing.T) {
	// pattern that would partial-match if not full
	re := regexp.MustCompile(`^adj:m:v_naz$`)
	m := &TokenMatcher{Entries: []MatcherEntry{{Lemma: "*", POS: re}}}
	pos := "adj:m:v_naz"
	tok := languagetool.NewAnalyzedToken("x", &pos, strPtr("x"))
	require.True(t, m.Matches(tok))
	pos2 := "adj:m:v_naz:compb"
	tok2 := languagetool.NewAnalyzedToken("x", &pos2, strPtr("x"))
	// MatchString("^adj:m:v_naz$") fails; full match fails — good
	require.False(t, m.Matches(tok2))
	// unanchored pattern still needs full-string via FindStringIndex
	re2 := regexp.MustCompile(`adj:m:v_naz`)
	m2 := &TokenMatcher{Entries: []MatcherEntry{{Lemma: "*", POS: re2}}}
	// "adj:m:v_naz:compb" contains substring but is not full match of unanchored re
	// FindStringIndex for "adj:m:v_naz" on longer string is [0:11] not full len
	require.False(t, m2.Matches(tok2))
	require.True(t, m2.Matches(tok))
}
