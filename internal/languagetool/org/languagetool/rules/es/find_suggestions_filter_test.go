package es

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrES(s string) *string { return &s }

func TestESFindSuggestionsRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.es.FindSuggestionsFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.es.FindSuggestionsFilter"))
}

func TestESFindSuggestionsAccept_Injected(t *testing.T) {
	ClearSpanishFilterSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingFromSimilarWords(func(token string) []string {
		if token == "csa" {
			return []string{"casa"}
		}
		return nil
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrES("NCFS000"), ptrES("casa")))
	}
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("csa", ptrES("UNKNOWN"), nil))
	tok.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("FS"), nil, 0, 3, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"casa"}, out.GetSuggestedReplacements())
}

func TestESFindSuggestionsNoSpeller(t *testing.T) {
	ClearSpanishFilterSpeller()
	f := NewFindSuggestionsFilter()
	// Default SpellingSuggestions empty without dict → no useful replacements
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("csa", ptrES("UNKNOWN"), nil))
	tok.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("FS"), nil, 0, 3, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
	// Missing tokens also nil
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, nil, nil))
}

func TestESFindSuggestions_WithESDict(t *testing.T) {
	ClearSpanishFilterSpeller()
	t.Cleanup(ClearSpanishFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Java FindSuggestionsFilter: /es/es-ES.dict; speller rule may use /es/hunspell/es.dict
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/es/src/main/resources/org/languagetool/resource/es/es-ES.dict"),
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/es/src/main/resources/org/languagetool/resource/es/hunspell/es.dict"),
		filepath.Join(root, "third_party/es/es-ES.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireSpanishFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("es-ES.dict / es.dict not in tree (Java FindSuggestionsFilter)")
	}
	f := NewFindSuggestionsFilter()
	// Soft POS: accept any suggestion (no SpanishTagger wire yet)
	f.MatchesDesiredPostag = func(suggestion, desiredPostag string) bool { return true }
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("csa", ptrES("UNKNOWN"), nil))
	tok.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("FS"), nil, 0, 3, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	// Dict may or may not suggest casa for csa; at least Accept should run without invent
	_ = out
	// Known: FilterDictSuggest returns something for a misspelled form when dict works
	sugg := FilterDictSuggest("csa")
	if len(sugg) > 0 {
		require.NotNil(t, out)
		require.NotEmpty(t, out.GetSuggestedReplacements())
	}
}
