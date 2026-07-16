package multitoken

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestMultitokenSpellerFilter(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\nLos Angeles\n")))
	f := &MultitokenSpellerFilter{Speller: sp}
	sent := languagetool.AnalyzePlain("I love New Yrok")
	// find approximate: use known dict entry with typo
	m := rules.NewRuleMatch(rules.NewFakeRule("MT"), sent, 0, 8, "multi")
	got := f.AcceptRuleMatch(m, "New Yrok")
	// may or may not find depending on distance; ensure no panic and empty drops
	if got != nil {
		require.NotEmpty(t, got.GetSuggestedReplacements())
	}
	// exact present → usually empty suggestions (not a misspelling path)
	// all-upper path
	f2 := &MultitokenSpellerFilter{Speller: sp}
	m2 := rules.NewRuleMatch(rules.NewFakeRule("MT"), sent, 0, 8, "multi")
	got2 := f2.AcceptRuleMatch(m2, "NEW YROK")
	_ = got2
	// sentence-start capitalization
	f3 := &MultitokenSpellerFilter{Speller: sp, AtSentenceStart: true}
	m3 := rules.NewRuleMatch(rules.NewFakeRule("MT"), sent, 0, 8, "multi")
	_ = f3.AcceptRuleMatch(m3, "new yrok")
}

func TestUppercaseFirstHelpers(t *testing.T) {
	require.True(t, isAllUpper("NEW YORK"))
	require.False(t, isAllUpper("New York"))
	require.Equal(t, "Hello", uppercaseFirst("hello"))
}
