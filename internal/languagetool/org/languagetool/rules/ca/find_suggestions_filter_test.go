package ca

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrFSCA(s string) *string { return &s }

func TestFindSuggestionsRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.FindSuggestionsFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.ca.FindSuggestionsFilter"))
}

func TestCAPreProcessWrongWord_ElaGeminada(t *testing.T) {
	require.Equal(t, "col·legi", caPreProcessWrongWord("col.legi"))
	require.Equal(t, "paral·lel", caPreProcessWrongWord("paral-lel"))
}

func TestCAIsSuggestionException(t *testing.T) {
	// enterar is ignored unless also enter
	atr := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("enterat", ptrFSCA("VMP00SM0"), ptrFSCA("enterar")))
	require.True(t, caIsSuggestionException(atr))
	// sentir is allowed
	atr2 := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("sentit", ptrFSCA("VMP00SM0"), ptrFSCA("sentir")))
	require.False(t, caIsSuggestionException(atr2))
}

func TestFindSuggestionsFilter_Accept(t *testing.T) {
	ClearCatalanFilterSpeller()
	f := NewFindSuggestionsFilter()
	f.SetSpellingSuggestions(func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"casa", "enterat"} // enterat from enterar should be exception
	})
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		switch word {
		case "casa":
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFSCA("NCFS000"), ptrFSCA("casa")))
		case "enterat":
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFSCA("VMP00SM0"), ptrFSCA("enterar")))
		default:
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFSCA("NCFS000"), nil))
		}
	}
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("csa", ptrFSCA("NCFS000"), nil))
	tok.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("FS"), nil, 0, 3, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"casa"}, out.GetSuggestedReplacements())
}

func TestFindSuggestionsFilter_NoSpeller(t *testing.T) {
	ClearCatalanFilterSpeller()
	f := NewFindSuggestionsFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("FS"), nil, 0, 1, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, nil, nil))
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("csa", ptrFSCA("NCFS000"), nil))
	tok.SetStartPos(0)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
}

func TestFindSuggestionsFilter_WithCADict(t *testing.T) {
	ClearCatalanFilterSpeller()
	t.Cleanup(ClearCatalanFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Java FindSuggestionsFilter: /ca/ca-ES_spelling.dict
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/ca/src/main/resources/org/languagetool/resource/ca/ca-ES_spelling.dict"),
		filepath.Join(root, "third_party/ca/ca-ES_spelling.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireCatalanFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("ca-ES_spelling.dict not in tree (Java FindSuggestionsFilter)")
	}
	f := NewFindSuggestionsFilter()
	f.MatchesDesiredPostag = func(suggestion, desiredPostag string) bool { return true }
	tok := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("csa", ptrFSCA("NCFS000"), nil))
	tok.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("FS"), nil, 0, 3, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"wordFrom": "1", "desiredPostag": "N.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	if len(FilterDictSuggest("csa")) > 0 {
		require.NotNil(t, out)
		require.NotEmpty(t, out.GetSuggestedReplacements())
	}
}
