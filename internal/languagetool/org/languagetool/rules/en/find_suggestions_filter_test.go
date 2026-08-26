package en

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrFS(s string) *string { return &s }

func fsTok(surface string, start int, pos *string) *languagetool.AnalyzedTokenReadings {
	atr := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(surface, pos, nil))
	atr.SetStartPos(start)
	return atr
}

func englishDictPath(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	return filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/english.dict")
}

func wireEnglishDictForFindSuggestions(t *testing.T) {
	t.Helper()
	ClearEnglishFilterSpeller()
	ClearEnglishFilterTagger()
	t.Cleanup(func() {
		ClearEnglishFilterSpeller()
		ClearEnglishFilterTagger()
	})
	pos := englishDictPath(t)
	if st, err := os.Stat(pos); err != nil || st.IsDir() {
		t.Skipf("english.dict missing: %s", pos)
	}
	// Java: same /en/english.dict for MorfologikSpeller + EnglishTagger.INSTANCE
	require.True(t, WireEnglishFilterTagger(pos), "open english.dict for tagger+findSimilarWords")
	require.True(t, FilterFindSuggestionsAvailable())
	require.True(t, FilterTaggerAvailable())
}

func TestFindSuggestionsFilter_WiresTagNoSynthesize(t *testing.T) {
	ClearEnglishFilterTagger()
	f := NewFindSuggestionsFilter()
	require.NotNil(t, f.Tag)
	// Java FindSuggestionsFilter does not override getSynthesizer() → null
	require.Nil(t, f.Synthesize)
	// Without WireEnglishFilterTagger, Tag returns nil (fail-closed)
	require.Nil(t, f.Tag("house"))
}

func TestFindSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.FindSuggestionsFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.en.FindSuggestionsFilter")
	require.NotNil(t, f)
	fs, ok := f.(*FindSuggestionsFilter)
	require.True(t, ok)
	require.NotNil(t, fs.Tag)
	require.Nil(t, fs.Synthesize) // Java getSynthesizer default null
	require.NotNil(t, fs.SpellingSuggestions)
}

func TestFindSuggestionsFilter_NoDictFailClosed(t *testing.T) {
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	ClearEnglishFilterTagger()
	f := NewFindSuggestionsFilter()
	tok := fsTok("runing", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 6, "msg")
	// no speller → Accept returns nil (fail-closed; do not invent similar words)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
	// default SpellingSuggestions also fail-closed
	require.Nil(t, f.SpellingSuggestions(tok))
}

func TestFindSuggestionsFilter_InjectedDesiredPostagFilter(t *testing.T) {
	// Java: keep only suggestions whose tagger readings match desiredPostag.
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"running", "ruins", "xxxnotaword"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		switch word {
		case "running":
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("VBG"), ptrFS("run")))
		case "ruins":
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("NNS"), ptrFS("ruin")))
		case "runing":
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("UNKNOWN"), nil))
		default:
			// untagged — MatchesPosTagRegex false
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, nil, nil))
		}
	}
	tok := fsTok("runing", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 6, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"running"}, out.GetSuggestedReplacements())
}

func TestFindSuggestionsFilter_DiacriticsModeDropWhenOriginalMatches(t *testing.T) {
	// Java: if original (pre-disambig tag) matches desiredPostag && Mode=diacritics → null
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"café"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		// original "cafe" already tags as noun
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("NN"), nil))
	}
	tok := fsTok("cafe", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 4, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*", "Mode": "diacritics",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
}

func TestFindSuggestionsFilter_DiacriticsModeKeepsOnlyDiacriticVariants(t *testing.T) {
	// Java: diacriticsMode keeps only equalWithoutDiacritics suggestions; drops unrelated.
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"café", "cafe", "boat"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		// original "cafe" must NOT match desiredPostag (else diacritics drops whole match)
		if word == "cafe" {
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("UNKNOWN"), nil))
		}
		if word == "boat" {
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("NN"), nil))
		}
		// café
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("NN"), nil))
	}
	tok := fsTok("cafe", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*", "Mode": "diacritics",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	// "cafe" equals original → excluded by suggestion != atrWord.getToken()
	// "boat" fails equalWithoutDiacritics
	require.Equal(t, []string{"café"}, out.GetSuggestedReplacements())
}

func TestFindSuggestionsFilter_DiacriticsEmptyDrops(t *testing.T) {
	// Java: diacriticsMode && replacements empty && no finished suggestion → null
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"boat"} // not a diacritic variant
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		if word == "cafe" {
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("UNKNOWN"), nil))
		}
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("NN"), nil))
	}
	tok := fsTok("cafe", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 4, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*", "Mode": "diacritics",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
}

func TestFindSuggestionsFilter_WordFromNumeric(t *testing.T) {
	// Java getPosition("2") → patternTokens[1]
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		require.Equal(t, "runing", atr.GetToken())
		return []string{"running"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		if word == "running" || word == "runing" {
			pos := "VBG"
			if word == "runing" {
				pos = "UNKNOWN"
			}
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS(pos), nil))
		}
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("DT"), nil))
	}
	toks := []*languagetool.AnalyzedTokenReadings{
		fsTok("the", 0, ptrFS("DT")),
		fsTok("runing", 4, ptrFS("UNKNOWN")),
	}
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 4, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "2", "desiredPostag": "VB.*",
	}, 0, toks, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"running"}, out.GetSuggestedReplacements())
}

func TestFindSuggestionsFilter_WordFromMarker(t *testing.T) {
	// Java getPosition("marker") → first pattern token at/after match FromPos
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		require.Equal(t, "runing", atr.GetToken())
		return []string{"running"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		if word == "running" {
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("VBG"), nil))
		}
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("UNKNOWN"), nil))
	}
	toks := []*languagetool.AnalyzedTokenReadings{
		fsTok("the", 0, ptrFS("DT")),
		fsTok("runing", 4, ptrFS("UNKNOWN")),
	}
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 4, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "marker", "desiredPostag": "VB.*",
	}, 0, toks, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"running"}, out.GetSuggestedReplacements())
}

func TestFindSuggestionsFilter_WordFromMarkerPlusOne(t *testing.T) {
	// Java getPosition("marker+1"): marker walk then +1 offset.
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		require.Equal(t, "runing", atr.GetToken())
		return []string{"running"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		if word == "running" {
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("VBG"), nil))
		}
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("UNKNOWN"), nil))
	}
	// match FromPos on "the"; marker → "the", marker+1 → "runing"
	toks := []*languagetool.AnalyzedTokenReadings{
		fsTok("the", 0, ptrFS("DT")),
		fsTok("runing", 4, ptrFS("UNKNOWN")),
	}
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 3, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "marker+1", "desiredPostag": "VB.*",
	}, 0, toks, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"running"}, out.GetSuggestedReplacements())
}

func TestFindSuggestionsFilter_WordFromOOBPanics(t *testing.T) {
	// Java getPosition OOB → IllegalArgumentException (panic in Go), not soft nil.
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"running"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("VBG"), nil))
	}
	tok := fsTok("runing", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 6, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"wordFrom": "9", "desiredPostag": "VB.*",
		}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	})
}

func TestFindSuggestionsFilter_OriginalErrorStrRemoveOnly(t *testing.T) {
	// Java: definitiveReplacements.remove(match.getOriginalErrorStr()) — no invent set.
	// Pre-seed non-template match suggestions so remove is observable.
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"running"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("VBG"), nil))
	}
	tok := fsTok("runing", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 6, "msg")
	m.SetSuggestedReplacements([]string{"runing", "also"})
	// OriginalErrorStr empty: remove("") must not invent setOriginalErrorStr from span
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "runing")
	require.Contains(t, out.GetSuggestedReplacements(), "also")
	require.Equal(t, "", m.GetOriginalErrorStr())

	m2 := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 6, "msg")
	m2.OriginalErrorStr = "runing"
	m2.SetSuggestedReplacements([]string{"runing", "also"})
	out2 := f.AcceptRuleMatch(m2, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out2)
	require.NotContains(t, out2.GetSuggestedReplacements(), "runing")
	require.Contains(t, out2.GetSuggestedReplacements(), "also")
	require.Contains(t, out2.GetSuggestedReplacements(), "running")
}

func TestFindSuggestionsFilter_SuppressMatchEmpty(t *testing.T) {
	// Java: suppressMatch=true and no replacements → null
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		return nil
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("UNKNOWN"), nil))
	}
	tok := fsTok("xyzzy", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "NN.*", "suppressMatch": "true",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
}

func TestFindSuggestionsFilter_SuggestionTemplate(t *testing.T) {
	// Java: "{suggestion}" placeholder expansion
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"running"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		if word == "running" {
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("VBG"), nil))
		}
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("UNKNOWN"), nil))
	}
	tok := fsTok("runing", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 6, "msg")
	m.SetSuggestedReplacements([]string{"try {suggestion}"})
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"try running"}, out.GetSuggestedReplacements())
}

func TestFindSuggestionsFilter_NoSynthesizePath(t *testing.T) {
	// Java getSynthesizer() null → replacements2 never filled; only spelling path.
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	f := NewFindSuggestionsFilter()
	require.Nil(t, f.Synthesize)
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		// candidate tags as NNS, not VB.* — without synth, no replacement
		return []string{"houses"}
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		if word == "houses" {
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFS("NNS"), ptrFS("house")))
		}
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFS("UNKNOWN"), nil))
	}
	tok := fsTok("houss", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*", "suppressMatch": "true",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	// no spelling match on VB.* and no synth → suppress drops
	require.Nil(t, out)
}

func TestFilterTagWord_WithDict(t *testing.T) {
	wireEnglishDictForFindSuggestions(t)
	atr := FilterTagWord("houses")
	require.NotNil(t, atr)
	// dict should yield NNS/NN-style readings
	require.True(t, atr.IsTagged() || atr.MatchesPosTagRegex("NN.*"),
		"expected tagged houses, got readings=%v", atr.GetReadings())
	require.True(t, FilterSuggestionMatchesPostag("houses", "NNS|NN"))
}

func TestFindSuggestionsFilter_WithEnglishDict_FindSimilarWords(t *testing.T) {
	// Official /en/english.dict + findSimilarWords + EnglishTagger desiredPostag filter.
	wireEnglishDictForFindSuggestions(t)
	f := NewFindSuggestionsFilter()
	// SpellingSuggestions must use real FindSimilarWords (not invent)
	sims := FilterDictSuggest("runing")
	require.NotEmpty(t, sims, "english.dict findSimilarWords(runing) must yield candidates")
	require.Contains(t, sims, "running")

	tok := fsTok("runing", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 6, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*|NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Contains(t, sugs, "running")
	// every surface must match desiredPostag via EnglishTagger-equivalent
	for _, s := range sugs {
		require.True(t, FilterSuggestionMatchesPostag(s, "VB.*|NN.*"),
			"suggestion %q must match desiredPostag", s)
	}
}

func TestFindSuggestionsFilter_WithEnglishDict_Recieve(t *testing.T) {
	wireEnglishDictForFindSuggestions(t)
	sims := FilterDictSuggest("recieve")
	require.Contains(t, sims, "receive")
	f := NewFindSuggestionsFilter()
	tok := fsTok("recieve", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 7, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*|NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "receive")
}

func TestFindSuggestionsFilter_ProcessWideWireFromTagger(t *testing.T) {
	// WireEnglishFilterTagger alone must enable FindSuggestions (Java same english.dict).
	ClearEnglishFilterSpeller()
	ClearEnglishFilterTagger()
	t.Cleanup(func() {
		ClearEnglishFilterSpeller()
		ClearEnglishFilterTagger()
	})
	pos := englishDictPath(t)
	if st, err := os.Stat(pos); err != nil || st.IsDir() {
		t.Skipf("english.dict missing: %s", pos)
	}
	require.False(t, FilterFindSuggestionsAvailable())
	require.True(t, WireEnglishFilterTagger(pos))
	require.True(t, FilterFindSuggestionsAvailable(), "tagger wire must install findSimilarWords dict")
	require.Contains(t, FilterDictSuggest("seperate"), "separate")
}

func TestFindSuggestionsFilter_NoFallbackToLocaleHunspell(t *testing.T) {
	// Java FindSuggestionsFilter uses only /en/english.dict — never locale en_US hunspell.
	// Wiring filterDict alone must leave FindSuggestions unavailable / fail-closed.
	ClearEnglishFilterSpeller()
	ClearEnglishFindSuggestionsSpeller()
	ClearEnglishFilterTagger()
	t.Cleanup(func() {
		ClearEnglishFilterSpeller()
		ClearEnglishFindSuggestionsSpeller()
		ClearEnglishFilterTagger()
	})
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	enUS := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	if st, err := os.Stat(enUS); err != nil || st.IsDir() {
		t.Skipf("en_US.dict missing: %s", enUS)
	}
	require.True(t, WireEnglishFilterSpeller(enUS))
	require.True(t, FilterDictAvailable(), "locale misspell dict wired")
	require.False(t, FilterFindSuggestionsAvailable(),
		"must not treat en_US.dict as FindSuggestions english.dict")
	require.Nil(t, FilterDictSuggest("runing"),
		"FilterDictSuggest must not fall back to locale hunspell")
	f := NewFindSuggestionsFilter()
	tok := fsTok("runing", 0, ptrFS("UNKNOWN"))
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 6, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "VB.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
}
