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

// Twin of MultitokenSpellerFilter.isMisspelled: WordTokenizer per token, not whole string.
func TestMultitokenSpellerFilter_IsMisspelledTokenizes(t *testing.T) {
	// Whole phrase "New York" is not in the single-token invent list; tokens "New" and "York" are OK.
	// "Nwe York" has misspelled "Nwe".
	seen := []string{}
	f := &MultitokenSpellerFilter{
		IsMisspelled: func(tok string) bool {
			seen = append(seen, tok)
			return tok == "Nwe" || tok == "xyz"
		},
	}
	require.False(t, f.isMisspelled("New York"), "known tokens → not misspelled")
	require.Contains(t, seen, "New")
	require.Contains(t, seen, "York")

	seen = nil
	require.True(t, f.isMisspelled("Nwe York"), "any bad token → misspelled")
	require.Contains(t, seen, "Nwe")

	// null speller
	fNil := &MultitokenSpellerFilter{}
	require.False(t, fNil.isMisspelled("anything"))
}

// Twin of areTokensAcceptedBySpeller language gate (en/de/pt/nl vs others).
func TestMultitokenSpellerFilter_AcceptedBySpellerGate(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	sent := languagetool.AnalyzePlain("New Yrok")
	m := rules.NewRuleMatch(rules.NewFakeRule("MT"), sent, 0, 8, "multi")

	// No CheckSpelling / IsMisspelled → Java non-en path: acceptedBySpeller=false
	f := &MultitokenSpellerFilter{Speller: sp}
	// Still may suggest; just ensure path does not panic
	_ = f.AcceptRuleMatch(m, "New Yrok")

	// Wired speller with all tokens "accepted" (none misspelled) — en/de/pt/nl path
	calls := 0
	f2 := &MultitokenSpellerFilter{
		Speller:       sp,
		CheckSpelling: true,
		IsMisspelled: func(tok string) bool {
			calls++
			return false // all tokens known
		},
	}
	// acceptedBySpeller = !false = true when CheckSpelling
	_ = f2.AcceptRuleMatch(m, "New Yrok")
	require.Greater(t, calls, 0, "must tokenize and probe tokens")

	// One bad token → acceptedBySpeller=false
	f3 := &MultitokenSpellerFilter{
		Speller:       sp,
		CheckSpelling: true,
		IsMisspelled:  func(tok string) bool { return tok == "Yrok" },
	}
	_ = f3.AcceptRuleMatch(m, "New Yrok")

	// IsMisspelled without CheckSpelling (fr/es/ca) must not probe tokens
	calls = 0
	f4 := &MultitokenSpellerFilter{
		Speller: sp,
		IsMisspelled: func(tok string) bool {
			calls++
			return true
		},
	}
	_ = f4.AcceptRuleMatch(m, "New Yrok")
	require.Equal(t, 0, calls, "non-en/de/pt/nl must leave acceptedBySpeller=false without probing")
}
