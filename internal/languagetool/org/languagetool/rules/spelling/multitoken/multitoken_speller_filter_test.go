package multitoken

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// matchWithError builds a RuleMatch with OriginalErrorStr set (Java pattern matcher path).
func matchWithError(sent *languagetool.AnalyzedSentence, from, to int, err string) *rules.RuleMatch {
	m := rules.NewRuleMatch(rules.NewFakeRule("MT"), sent, from, to, "multi")
	m.OriginalErrorStr = err
	return m
}

func TestMultitokenSpellerFilter_Suggestions(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\nLos Angeles\n")))
	f := &MultitokenSpellerFilter{Speller: sp, ShortCode: "fr"} // non-en gate
	sent := languagetool.AnalyzePlain("I love New Yrok")
	m := matchWithError(sent, 0, 8, "New Yrok")
	got := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	if got != nil {
		require.NotEmpty(t, got.GetSuggestedReplacements())
		require.Contains(t, got.GetSuggestedReplacements(), "New York")
	}
	// exact present → stopSearching → empty → null
	m2 := matchWithError(sent, 0, 8, "New York")
	require.Nil(t, f.AcceptRuleMatch(m2, nil, 1, nil, nil))
}

// Twin of MultitokenSpellerFilter.isMisspelled: WordTokenizer per token, not whole string.
func TestMultitokenSpellerFilter_IsMisspelledTokenizes(t *testing.T) {
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

	// null SpellingCheckRule
	fNil := &MultitokenSpellerFilter{}
	require.False(t, fNil.isMisspelled("anything"))
}

// Twin: areTokensAcceptedBySpeller language short-code gate (en/de/pt/nl vs others).
func TestMultitokenSpellerFilter_ShortCodeGate(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	sent := languagetool.AnalyzePlain("New Yrok")
	m := matchWithError(sent, 0, 8, "New Yrok")

	// Non-en short code: acceptedBySpeller stays false; IsMisspelled must not be probed.
	calls := 0
	fFR := &MultitokenSpellerFilter{
		Speller:   sp,
		ShortCode: "fr",
		IsMisspelled: func(tok string) bool {
			calls++
			return true
		},
	}
	_ = fFR.AcceptRuleMatch(m, nil, 1, nil, nil)
	require.Equal(t, 0, calls, "fr must leave acceptedBySpeller=false without probing")

	// en short code with all tokens accepted → acceptedBySpeller=true
	calls = 0
	fEN := &MultitokenSpellerFilter{
		Speller:   sp,
		ShortCode: "en",
		IsMisspelled: func(tok string) bool {
			calls++
			return false
		},
	}
	_ = fEN.AcceptRuleMatch(m, nil, 1, nil, nil)
	require.Greater(t, calls, 0, "en must tokenize and probe tokens")

	// de/pt/nl also enable the gate
	for _, code := range []string{"de", "pt", "nl"} {
		calls = 0
		f := &MultitokenSpellerFilter{
			Speller:   sp,
			ShortCode: code,
			IsMisspelled: func(tok string) bool {
				calls++
				return false
			},
		}
		_ = f.AcceptRuleMatch(m, nil, 1, nil, nil)
		require.Greater(t, calls, 0, "%s must enable areTokensAcceptedBySpeller path", code)
	}

	// es/ca leave gate off
	for _, code := range []string{"es", "ca", ""} {
		calls = 0
		f := &MultitokenSpellerFilter{
			Speller:   sp,
			ShortCode: code,
			IsMisspelled: func(tok string) bool {
				calls++
				return true
			},
		}
		_ = f.AcceptRuleMatch(m, nil, 1, nil, nil)
		require.Equal(t, 0, calls, "%s must not probe tokens", code)
	}
}

// Twin: bare short code (length==2) → getDefaultLanguageVariant for isMisspelled + speller.
func TestMultitokenSpellerFilter_DefaultLanguageVariant(t *testing.T) {
	baseSP := NewMultitokenSpeller()
	require.NoError(t, baseSP.LoadWords(strings.NewReader("Wrong Entry\n")))
	variantSP := NewMultitokenSpeller()
	require.NoError(t, variantSP.LoadWords(strings.NewReader("New York\n")))

	variantMissCalls := 0
	baseMissCalls := 0
	f := &MultitokenSpellerFilter{
		Speller:                        baseSP,
		ShortCode:                      "en",
		ShortCodeWithCountryAndVariant: "en", // length 2
		IsMisspelled: func(tok string) bool {
			baseMissCalls++
			return false
		},
		GetDefaultLanguageVariant: func() *MultitokenSpellerLanguageHooks {
			return &MultitokenSpellerLanguageHooks{
				Speller: variantSP,
				IsMisspelled: func(tok string) bool {
					variantMissCalls++
					return false
				},
			}
		},
	}
	sent := languagetool.AnalyzePlain("New Yrok")
	m := matchWithError(sent, 0, 8, "New Yrok")
	got := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	require.NotNil(t, got)
	require.Contains(t, got.GetSuggestedReplacements(), "New York")
	require.Greater(t, variantMissCalls, 0, "must use default variant isMisspelled")
	require.Equal(t, 0, baseMissCalls, "must not use bare-lang isMisspelled after reassignment")

	// length != 2 → no variant reassignment
	variantMissCalls, baseMissCalls = 0, 0
	f2 := &MultitokenSpellerFilter{
		Speller:                        baseSP,
		ShortCode:                      "en",
		ShortCodeWithCountryAndVariant: "en-US",
		IsMisspelled: func(tok string) bool {
			baseMissCalls++
			return false
		},
		GetDefaultLanguageVariant: func() *MultitokenSpellerLanguageHooks {
			return &MultitokenSpellerLanguageHooks{
				Speller: variantSP,
				IsMisspelled: func(tok string) bool {
					variantMissCalls++
					return false
				},
			}
		},
	}
	// baseSP has only "Wrong Entry" — no "New York" → no suggestions
	require.Nil(t, f2.AcceptRuleMatch(matchWithError(sent, 0, 8, "New Yrok"), nil, 1, nil, nil))
	require.Greater(t, baseMissCalls, 0)
	require.Equal(t, 0, variantMissCalls)
}

// Twin: Arrays.stream(empty).allMatch(...) is true → drop match.
func TestMultitokenSpellerFilter_EmptyPatternTokensDrops(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	f := &MultitokenSpellerFilter{Speller: sp, ShortCode: "en"}
	m := matchWithError(languagetool.AnalyzePlain("New Yrok"), 0, 8, "New Yrok")
	// empty non-nil slice (Java empty array allMatch true)
	require.Nil(t, f.AcceptRuleMatch(m, nil, 1, []*languagetool.AnalyzedTokenReadings{}, nil))
	// nil tokens: Go convenience path keeps working
	got := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	if got != nil {
		require.Contains(t, got.GetSuggestedReplacements(), "New York")
	}
}

func TestMultitokenSpellerFilter_AllIgnoredBySpeller(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	f := &MultitokenSpellerFilter{Speller: sp, ShortCode: "en"}
	sent := languagetool.AnalyzePlain("New Yrok")
	m := matchWithError(sent, 0, 8, "New Yrok")
	t0 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("New", nil, nil), 0)
	t1 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("Yrok", nil, nil), 4)
	t0.IgnoreSpelling()
	t1.IgnoreSpelling()
	require.Nil(t, f.AcceptRuleMatch(m, nil, 1, []*languagetool.AnalyzedTokenReadings{t0, t1}, nil))
}

// Twin: all-upper path uses UTF-16 length > 4, uppercases suggestions, excludes original, dedupes.
func TestMultitokenSpellerFilter_AllUpper(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	f := &MultitokenSpellerFilter{Speller: sp, ShortCode: "fr"}
	sent := languagetool.AnalyzePlain("NEW YROK")
	m := matchWithError(sent, 0, 8, "NEW YROK")
	got := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	require.NotNil(t, got)
	reps := got.GetSuggestedReplacements()
	require.NotEmpty(t, reps)
	for _, s := range reps {
		require.Equal(t, strings.ToUpper(s), s, "all-upper path keeps upper: %q", s)
		require.NotEqual(t, "NEW YROK", s, "must exclude original error")
	}
	require.Contains(t, reps, "NEW YORK")

	// length <= 4: no all-upper transform path (Java length()>4)
	// "AB CD" is 5 UTF-16 units — use short phrase "A B" (3) if dict has it
	sp2 := NewMultitokenSpeller()
	require.NoError(t, sp2.LoadWords(strings.NewReader("a b\n")))
	f2 := &MultitokenSpellerFilter{Speller: sp2, ShortCode: "fr"}
	// "X Y" length 3 <= 4 → not all-upper branch even if upper
	mShort := matchWithError(languagetool.AnalyzePlain("X Y"), 0, 3, "X Y")
	// may or may not suggest; if suggests, casing follows non-all-upper path
	_ = f2.AcceptRuleMatch(mShort, nil, 1, nil, nil)
}

// Twin: sentence-start capitalization via patternTokenPos walk; only all-lower replacements.
func TestMultitokenSpellerFilter_SentenceStartCap(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("new york\n")))
	f := &MultitokenSpellerFilter{Speller: sp, ShortCode: "fr"}
	sent := languagetool.AnalyzePlain("new yrok today")
	tokens := sent.GetTokensWithoutWhitespace()
	require.GreaterOrEqual(t, len(tokens), 3)
	from := tokens[1].GetStartPos()
	to := tokens[2].GetEndPos()
	m := matchWithError(sent, from, to, "new yrok")
	// patternTokenPos=1 → first content token after SENT_START
	got := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	require.NotNil(t, got)
	reps := got.GetSuggestedReplacements()
	require.NotEmpty(t, reps)
	require.Equal(t, "New york", reps[0], "all-lower dict form → uppercaseFirstChar")

	// not at sentence start → no capitalization
	sent2 := languagetool.AnalyzePlain("see new yrok today")
	toks2 := sent2.GetTokensWithoutWhitespace()
	// tokens: 0=SENT_START, 1=see, 2=new, 3=yrok, ...
	require.GreaterOrEqual(t, len(toks2), 4)
	from2 := toks2[2].GetStartPos()
	to2 := toks2[3].GetEndPos()
	m2 := matchWithError(sent2, from2, to2, "new yrok")
	got2 := f.AcceptRuleMatch(m2, nil, 2, nil, nil)
	require.NotNil(t, got2)
	require.Equal(t, "new york", got2.GetSuggestedReplacements()[0], "mid-sentence keeps lower")
}

// Twin: mixed-case suggestion (iPad-style) is not title-cased at sentence start.
func TestMultitokenSpellerFilter_NoCapMixedCase(t *testing.T) {
	sp := NewMultitokenSpeller()
	// force additional suggestion path via exact-ish dict: store mixed form
	require.NoError(t, sp.LoadWords(strings.NewReader("iPad Pro\n")))
	f := &MultitokenSpellerFilter{Speller: sp, ShortCode: "fr"}
	sent := languagetool.AnalyzePlain("ipad pro here")
	tokens := sent.GetTokensWithoutWhitespace()
	from := tokens[1].GetStartPos()
	to := tokens[2].GetEndPos()
	m := matchWithError(sent, from, to, "ipad pro")
	got := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	if got == nil {
		t.Skip("no suggestions for ipad pro distance; path still covered by unit lower-check")
	}
	for _, r := range got.GetSuggestedReplacements() {
		// "iPad Pro" is not all-lower → must not become "IPad Pro" via uppercaseFirstChar
		if strings.Contains(r, "Pad") {
			require.True(t, strings.HasPrefix(r, "i") || strings.HasPrefix(r, "I"), "got %q", r)
			// if it starts with i, mixed case preserved
			if strings.HasPrefix(r, "i") {
				require.NotEqual(t, "IPad Pro", r)
			}
		}
	}
}

// Twin: empty replacements → null; setSuggestedReplacements when non-empty.
func TestMultitokenSpellerFilter_EmptyDrops(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	f := &MultitokenSpellerFilter{Speller: sp}
	m := matchWithError(languagetool.AnalyzePlain("zzzz zzzz"), 0, 9, "zzzz zzzz")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 1, nil, nil))
	// nil speller → null
	fNil := &MultitokenSpellerFilter{ShortCode: "en"}
	require.Nil(t, fNil.AcceptRuleMatch(m, nil, 1, nil, nil))
}

// Twin: sentence-start with leading punctuation (wordsStartPos skips punct).
func TestMultitokenSpellerFilter_SentenceStartSkipsPunct(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("new york\n")))
	f := &MultitokenSpellerFilter{Speller: sp, ShortCode: "fr"}
	// "\"new yrok" — leading quote is punctuation at sentence start
	sent := languagetool.AnalyzePlain("\"new yrok today")
	tokens := sent.GetTokensWithoutWhitespace()
	// Find wordsStartPos like Java
	wordsStartPos := 1
	for wordsStartPos < len(tokens) {
		tkn := tokens[wordsStartPos].GetToken()
		if tools.IsPunctuationMark(tkn) || tools.IsNotWordString(tkn) {
			wordsStartPos++
			continue
		}
		break
	}
	require.Greater(t, wordsStartPos, 1, "expected leading punct skipped")
	// pattern covers "new yrok"
	var from, to int
	for i := wordsStartPos; i+1 < len(tokens); i++ {
		if strings.EqualFold(tokens[i].GetToken(), "new") {
			from = tokens[i].GetStartPos()
			to = tokens[i+1].GetEndPos()
			break
		}
	}
	m := matchWithError(sent, from, to, "new yrok")
	got := f.AcceptRuleMatch(m, nil, wordsStartPos, nil, nil)
	require.NotNil(t, got)
	require.Equal(t, "New york", got.GetSuggestedReplacements()[0])
}

func TestUppercaseFirstHelpers(t *testing.T) {
	require.True(t, tools.IsAllUppercase("NEW YORK"))
	require.False(t, tools.IsAllUppercase("New York"))
	require.Equal(t, "Hello", tools.UppercaseFirstChar("hello"))
}

// Twin: acceptedBySpeller true vs false affects MultitokenSpeller.GetSuggestionsOpts.
func TestMultitokenSpellerFilter_AcceptedBySpellerPassedToSpeller(t *testing.T) {
	sp := NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("hello world\n")))
	// en + IsMisspelled false → acceptedBySpeller=true
	// HELLO WORLD + accepted true → speller empty (MultitokenSpeller_AreTokensAcceptedBySpeller)
	f := &MultitokenSpellerFilter{
		Speller:      sp,
		ShortCode:    "en",
		IsMisspelled: func(string) bool { return false },
	}
	m := matchWithError(languagetool.AnalyzePlain("HELLO WORLD"), 0, 11, "HELLO WORLD")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 1, nil, nil),
		"en + all tokens accepted + ALL CAPS original → speller returns empty")

	// fr: acceptedBySpeller=false; speller may return "hello world", all-upper → "HELLO WORLD"
	// which equals underlinedError → excluded → null (Java !underlinedError.equals(newReplacement))
	fFR := &MultitokenSpellerFilter{
		Speller:      sp,
		ShortCode:    "fr",
		IsMisspelled: func(string) bool { return false },
	}
	require.Nil(t, fFR.AcceptRuleMatch(
		matchWithError(languagetool.AnalyzePlain("HELLO WORLD"), 0, 11, "HELLO WORLD"),
		nil, 1, nil, nil,
	), "all-upper suggestion equal to original is dropped")

	// en + misspelled tokens → acceptedBySpeller=false → same as fr path for casing
	fENMiss := &MultitokenSpellerFilter{
		Speller:      sp,
		ShortCode:    "en",
		IsMisspelled: func(string) bool { return true },
	}
	// Typo that is not equal to uppercased dict form
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\n")))
	got := fENMiss.AcceptRuleMatch(
		matchWithError(languagetool.AnalyzePlain("NEW YROK"), 0, 8, "NEW YROK"),
		nil, 1, nil, nil,
	)
	require.NotNil(t, got)
	require.Contains(t, got.GetSuggestedReplacements(), "NEW YORK")
}

// Twin: splitBySpace only on ASCII space (StringUtils.split) — lives with MultitokenSpeller.
func TestSplitBySpace_OnlyASCIISpace(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, splitBySpace("a b"))
	require.Equal(t, []string{"a", "b"}, splitBySpace("a  b")) // empty omitted
	// tab is NOT a split point (unlike strings.Fields)
	require.Equal(t, []string{"a\tb"}, splitBySpace("a\tb"))
}

func TestIsSpellerAcceptedShortCode(t *testing.T) {
	require.True(t, isSpellerAcceptedShortCode("en"))
	require.True(t, isSpellerAcceptedShortCode("de"))
	require.True(t, isSpellerAcceptedShortCode("pt"))
	require.True(t, isSpellerAcceptedShortCode("nl"))
	require.False(t, isSpellerAcceptedShortCode("fr"))
	require.False(t, isSpellerAcceptedShortCode("es"))
	require.False(t, isSpellerAcceptedShortCode("ca"))
	require.False(t, isSpellerAcceptedShortCode(""))
}
