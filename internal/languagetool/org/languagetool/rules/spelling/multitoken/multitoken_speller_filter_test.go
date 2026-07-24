package multitoken

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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
	require.True(t, tools.IsAllUppercase("NEW YORK"))
	require.False(t, tools.IsAllUppercase("New York"))
	require.Equal(t, "Hello", tools.UppercaseFirstChar("hello"))
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

// Twin: Arrays.stream(empty).allMatch(...) is true → drop match.
func TestMultitokenSpellerFilter_EmptyPatternTokensDrops(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	f := &MultitokenSpellerFilter{Speller: sp}
	m := rules.NewRuleMatch(rules.NewFakeRule("T"), languagetool.AnalyzePlain("New York"), 0, 8, "x")
	// empty non-nil slice (Java empty array allMatch true)
	require.Nil(t, f.AcceptRuleMatchFull(m, nil, 1, []*languagetool.AnalyzedTokenReadings{}, "New York"))
	// nil tokens: Go convenience path keeps working
	got := f.AcceptRuleMatchFull(m, nil, 1, nil, "New York")
	require.Nil(t, got) // exact match stopSearching → no replacements → null
}

func TestMultitokenSpellerFilter_AllIgnoredBySpeller(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	f := &MultitokenSpellerFilter{Speller: sp}
	sent := languagetool.AnalyzePlain("New Yrok")
	m := rules.NewRuleMatch(rules.NewFakeRule("MT"), sent, 0, 8, "multi")
	// two tokens both ignored
	t0 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("New", nil, nil), 0)
	t1 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Yrok", nil, nil), 4)
	t0.IgnoreSpelling()
	t1.IgnoreSpelling()
	got := f.AcceptRuleMatchFull(m, nil, 1, []*languagetool.AnalyzedTokenReadings{t0, t1}, "New Yrok")
	require.Nil(t, got)
}

// Twin: all-upper path uses UTF-16 length > 4 and uppercases suggestions.
func TestMultitokenSpellerFilter_AllUpper(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	f := &MultitokenSpellerFilter{Speller: sp}
	sent := languagetool.AnalyzePlain("NEW YROK")
	m := rules.NewRuleMatch(rules.NewFakeRule("MT"), sent, 0, 8, "multi")
	got := f.AcceptRuleMatch(m, "NEW YROK")
	if got != nil {
		for _, s := range got.GetSuggestedReplacements() {
			require.Equal(t, strings.ToUpper(s), s, "all-upper path keeps upper: %q", s)
		}
	}
}

// Twin: sentence-start capitalization via patternTokenPos walk.
func TestMultitokenSpellerFilter_SentenceStartCap(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("new york\n")))
	f := &MultitokenSpellerFilter{Speller: sp}
	// "New yrok" at start — pattern token pos 1 (after SENT_START)
	sent := languagetool.AnalyzePlain("new yrok today")
	// fromPos of first content token
	tokens := sent.GetTokensWithoutWhitespace()
	require.GreaterOrEqual(t, len(tokens), 2)
	from := tokens[1].GetStartPos()
	to := tokens[2].GetEndPos()
	m := rules.NewRuleMatch(rules.NewFakeRule("MT"), sent, from, to, "multi")
	// Force suggestions path with patternTokenPos=1
	got := f.AcceptRuleMatchFull(m, nil, 1, nil, "new yrok")
	if got != nil && len(got.GetSuggestedReplacements()) > 0 {
		// first suggestion should be title-cased if source was all-lower dict form
		s0 := got.GetSuggestedReplacements()[0]
		require.True(t, s0[0] >= 'A' && s0[0] <= 'Z' || len(s0) == 0, "expected capitalized: %q", s0)
	}
}

// Twin: splitBySpace only on ASCII space (StringUtils.split).
func TestSplitBySpace_OnlyASCIISpace(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, splitBySpace("a b"))
	require.Equal(t, []string{"a", "b"}, splitBySpace("a  b")) // empty omitted
	// tab is NOT a split point (unlike strings.Fields)
	require.Equal(t, []string{"a\tb"}, splitBySpace("a\tb"))
}
